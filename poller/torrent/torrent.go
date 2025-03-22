package torrent

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"net/http"

	"github.com/jackpal/bencode-go"
	"github.com/lonord/rss-torrent-downloader/poller"
)

type torrentPoller struct {
}

func init() {
	poller.RegisterPuller(&torrentPoller{})
}

func (p *torrentPoller) Poll(ctx context.Context, rss *poller.RSSItem, options map[string]string) (*poller.Job, bool, error) {
	if rss.Enclosure.Type != "application/x-bittorrent" {
		return nil, false, nil
	}
	torrentData, err := fetchTorrent(ctx, rss.Enclosure.URL)
	if err != nil {
		return nil, true, err
	}
	infoHash, err := calculateInfoHash(torrentData)
	if err != nil {
		return nil, true, err
	}
	return &poller.Job{
		Type:     "torrent",
		Content:  base64.StdEncoding.EncodeToString(torrentData),
		InfoHash: infoHash,
	}, true, nil
}

func calculateInfoHash(torrentContent []byte) (string, error) {
	torrentDataReader := bytes.NewReader(torrentContent)
	result, err := bencode.Decode(torrentDataReader)
	if err != nil {
		return "", err
	}
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return "", errors.New("invalid torrent file")
	}
	hashValue, ok := resultMap["hash"]
	if !ok {
		info, ok2 := resultMap["info"]
		if !ok2 {
			return "", errors.New("missing info field")
		}
		encodedInfo := &bytes.Buffer{}
		if err := bencode.Marshal(encodedInfo, info); err != nil {
			return "", err
		}
		hash := sha1.New()
		hash.Write(encodedInfo.Bytes())
		infoHash := hash.Sum(nil)
		hashValue = hex.EncodeToString(infoHash)
	}
	hashString, ok := hashValue.(string)
	if !ok {
		return "", errors.New("invalid hash field")
	}
	return hashString, nil
}

func fetchTorrent(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New("fetch torrent failed: " + res.Status)
	}
	return io.ReadAll(res.Body)
}
