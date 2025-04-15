package downloader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"path"

	"github.com/google/uuid"
	"github.com/lonord/rss-torrent-downloader/poller"
)

type Aria2Downloader struct {
	URL    string
	Secret string
	Dir    string
}

type RPCRequest struct {
	Version string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type RPCResponse struct {
	ID    string                 `json:"id"`
	Error map[string]interface{} `json:"error"`
}

type AddResponse struct {
	RPCResponse
	Result string `json:"result"`
}

type TellItem struct {
	GID             string `json:"gid"`
	Status          string `json:"status"`
	CompletedLength string `json:"completedLength"`
	TotalLength     string `json:"totalLength"`
	DownloadSpeed   string `json:"downloadSpeed"`
	InfoHash        string `json:"infoHash"`
}

type TellResponse struct {
	RPCResponse
	Result []TellItem `json:"result"`
}

func (d *Aria2Downloader) BatchDownload(ctx context.Context, works []*poller.Work) ([]DownloadResult, error) {
	items, err := d.tellAll(ctx)
	if err != nil {
		return nil, err
	}
	itemMap := make(map[string]*TellItem)
	for _, item := range items {
		if _, alreadyExist := itemMap[item.InfoHash]; !alreadyExist || item.Status == "active" || item.Status == "waiting" || item.Status == "paused" {
			itemMap[item.InfoHash] = &item
		}
	}
	results := make([]DownloadResult, len(works))
	for i, work := range works {
		downOpts := map[string]interface{}{
			"dir": path.Join(d.Dir, work.Name),
		}
		for k, v := range work.Aria2Opt {
			downOpts[k] = v
		}
		var r DownloadResult
		for _, job := range work.Jobs {
			item, ok := itemMap[job.InfoHash]
			if !ok || item.Status == "error" {
				if err := d.addTorrent(ctx, downOpts, job); err != nil {
					log.Printf("aria2: add torrent %s@%s failed: %v\n", job.InfoHash, work.Name, err)
					r.Failed++
				} else {
					log.Printf("aria2: add torrent %s@%s\n", job.InfoHash, work.Name)
					r.Added++
				}
			} else if item.Status == "complete" {
				// remove task from aria2 server
				if err := d.remove(ctx, item.GID); err != nil {
					log.Printf("remove task from aria2c error: %s, gid: %s, infoHash: %s\n", err, item.GID, item.InfoHash)
				} else {
					r.Completed = append(r.Completed, job.InfoHash)
				}
			} else if item.Status == "removed" {
				r.Removed = append(r.Removed, job.InfoHash)
			} else {
				r.Running++
			}
		}
		results[i] = r
	}
	return results, nil
}

func (d *Aria2Downloader) addTorrent(ctx context.Context, options map[string]interface{}, job *poller.Job) error {
	options["follow-torrent"] = "mem"
	options["seed-time"] = 0
	req := d.newReq("aria2.addTorrent", job.Content, []string{}, options)
	var resp AddResponse
	err := d.rpcCall(ctx, req, &resp)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return errors.New(resp.Error["message"].(string))
	}
	return nil
}

func (d *Aria2Downloader) remove(ctx context.Context, gid string) error {
	req := d.newReq("aria2.removeDownloadResult", gid)
	var resp AddResponse
	err := d.rpcCall(ctx, req, &resp)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return errors.New(resp.Error["message"].(string))
	}
	return nil
}

func (d *Aria2Downloader) tellAll(ctx context.Context) ([]TellItem, error) {
	columns := []string{"gid", "status", "completedLength", "totalLength", "downloadSpeed", "infoHash"}
	var items []TellItem
	if err := d.rpcCallTell(ctx, d.newReq("aria2.tellActive", columns), &items); err != nil {
		return nil, err
	}
	if err := d.rpcCallTell(ctx, d.newReq("aria2.tellWaiting", 0, 99999, columns), &items); err != nil {
		return nil, err
	}
	if err := d.rpcCallTell(ctx, d.newReq("aria2.tellStopped", 0, 99999, columns), &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (d *Aria2Downloader) rpcCallTell(ctx context.Context, rpc *RPCRequest, items *[]TellItem) error {
	var resp TellResponse
	err := d.rpcCall(ctx, rpc, &resp)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return errors.New(resp.Error["message"].(string))
	}
	for _, item := range resp.Result {
		if item.InfoHash != "" {
			*items = append(*items, item)
		}
	}
	return nil
}

func (d *Aria2Downloader) rpcCall(ctx context.Context, rpc *RPCRequest, out interface{}) error {
	body, err := json.Marshal(rpc)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", d.URL+"/jsonrpc", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err == nil {
			return errors.New("bad status code: " + resp.Status + ", result: " + string(b))
		} else {
			return errors.New("bad status code: " + resp.Status)
		}
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (d *Aria2Downloader) newReq(method string, params ...interface{}) *RPCRequest {
	params2 := params
	if d.Secret != "" {
		params2 = append([]interface{}{"token:" + d.Secret}, params...)
	}
	return &RPCRequest{
		Version: "2.0",
		ID:      uuid.New().String(),
		Method:  method,
		Params:  params2,
	}
}
