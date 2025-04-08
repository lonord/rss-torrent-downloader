package poller

import (
	"context"
	"encoding/xml"
	"errors"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

var pollers []Poller

type Poller interface {
	Poll(ctx context.Context, rss *RSSItem, options map[string]string) (*Job, bool, error)
}

type Work struct {
	Name     string
	Jobs     []*Job
	Aria2Opt map[string]string
}

func (w *Work) RemoveCompletedJob(completed []string) {
	for i := len(w.Jobs) - 1; i >= 0; i-- {
		job := w.Jobs[i]
		if slices.Contains(completed, job.InfoHash) {
			w.Jobs = slices.Delete(w.Jobs, i, i+1)
		}
	}
}

type Job struct {
	Type     string
	Content  string
	InfoHash string
}

type RSSWrapper struct {
	XMLName xml.Name `xml:"rss"`
	RSS     *RSS     `xml:"channel"`
}

type RSS struct {
	XMLName xml.Name   `xml:"channel"`
	Title   string     `xml:"title"`
	Link    string     `xml:"link"`
	Desc    string     `xml:"description"`
	Items   []*RSSItem `xml:"item"`
}

type RSSItem struct {
	Title     string       `xml:"title"`
	Link      string       `xml:"link"`
	Desc      string       `xml:"description"`
	Entry     TorrentEntry `xml:"torrent"`
	Enclosure Enclosure    `xml:"enclosure"`
}

type TorrentEntry struct {
	Link          string `xml:"link"`
	ContentLength uint64 `xml:"contentLength"`
	PubDate       string `xml:"pubDate"`
}

type Enclosure struct {
	URL    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

func RegisterPuller(poller Poller) {
	pollers = append(pollers, poller)
}

func Poll(ctx context.Context, rssURL string, options map[string]string) (*Work, error) {
	rss, err := fetchRSS(ctx, rssURL)
	if err != nil {
		return nil, err
	}
	w := &Work{
		Name:     strings.TrimSpace(rss.Title),
		Jobs:     []*Job{},
		Aria2Opt: make(map[string]string),
	}
	nameFilter, nameFilterEnable := options["filter"]
	timeFilter, timeFilterEnable := options["time"]
	sizeFilter, sizeFilterEnable := func() (uint64, bool) {
		s, ok := options["size"]
		if !ok {
			return 0, false
		}
		n, err := strconv.ParseUint(s, 10, 64)
		return n, err == nil
	}()
	for _, item := range rss.Items {
		if nameFilterEnable && !strings.Contains(item.Title, nameFilter) {
			continue
		}
		if timeFilterEnable && timeSmallerThan(item.Entry.PubDate, timeFilter) {
			continue
		}
		if sizeFilterEnable && item.Entry.ContentLength > sizeFilter {
			continue
		}
		job, err := pollItem(ctx, item, options)
		if err != nil {
			log.Printf("ignore poll failed item with error: %s, title: %s, type: %s, url: %s\n", err, item.Title, item.Enclosure.Type, item.Enclosure.URL)
			continue
		}
		w.Jobs = append(w.Jobs, job)
	}
	if trim, ok := options["trim"]; ok {
		w.Name = strings.TrimPrefix(w.Name, trim)
		w.Name = strings.TrimSuffix(w.Name, trim)
		w.Name = strings.TrimSpace(w.Name)
	}
	w.Name = formatFileName(w.Name)
	return w, nil
}

func fetchRSS(ctx context.Context, rssURL string) (*RSS, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", rssURL, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	decoder := xml.NewDecoder(res.Body)
	rssWrapper := &RSSWrapper{
		RSS: &RSS{},
	}
	if err := decoder.Decode(rssWrapper); err != nil {
		return nil, err
	}
	return rssWrapper.RSS, nil
}

func pollItem(ctx context.Context, item *RSSItem, options map[string]string) (*Job, error) {
	ctx2, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	for _, p := range pollers {
		if job, ok, err := p.Poll(ctx2, item, options); ok {
			if err != nil {
				return nil, err
			}
			return job, nil
		}
	}
	return nil, errors.New("poller not found")
}

func timeSmallerThan(t1, t2 string) bool {
	if t1 == "" || t2 == "" {
		return false
	}
	time1, err := parseTime(t1)
	if err != nil {
		return false
	}
	time2, err := parseTime(t2)
	if err != nil {
		return false
	}
	return time1.Compare(time2) < 0
}

func parseTime(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02T15:04:05", s)
	if err == nil {
		return t, nil
	}
	t, err = time.Parse("2006/01/02 15:04", s)
	if err == nil {
		return t, nil
	}
	return time.Time{}, errors.New("time format not supported: " + s)
}
