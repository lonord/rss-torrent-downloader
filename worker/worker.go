package worker

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/lonord/rss-torrent-downloader/downloader"
	"github.com/lonord/rss-torrent-downloader/poller"
	_ "github.com/lonord/rss-torrent-downloader/poller/torrent"
)

type Downloader interface {
	BatchDownload(ctx context.Context, works []*poller.Work) ([]downloader.DownloadResult, error)
}

type SubscriptionEntry struct {
	ID        string            `json:"-"`
	RssURL    string            `json:"url"`
	Options   map[string]string `json:"options"`
	Completed []string          `json:"completed"`
}

func (s *SubscriptionEntry) AddCompleted(completed []string) bool {
	changed := false
	for _, c := range completed {
		found := false
		for _, c0 := range s.Completed {
			if c0 == c {
				found = true
			}
		}
		if !found {
			s.Completed = append(s.Completed, c)
			changed = true
		}
	}
	return changed
}

type SubscriptionRepo interface {
	Query(fn func(entry *SubscriptionEntry)) error
	Save(entry *SubscriptionEntry) error
	Delete(id string) error
}

type Worker struct {
	Interval time.Duration
	Repo     SubscriptionRepo
	Down     Downloader

	mu sync.Mutex
}

func (w *Worker) Run() {
	for {
		w.doPoll()
		time.Sleep(w.Interval)
	}
}

func (w *Worker) doPoll() {
	w.mu.Lock()
	defer w.mu.Unlock()
	allCount := 0
	works := []*poller.Work{}
	entries := []*SubscriptionEntry{}
	w.Repo.Query(func(entry *SubscriptionEntry) {
		allCount++
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
		defer cancel()
		work, err := poller.Poll(ctx, entry.RssURL, entry.Options)
		if err != nil {
			log.Printf("poll %s error: %s\n", entry.RssURL, err)
			return
		}
		work.RemoveCompletedJob(entry.Completed)
		if len(work.Jobs) == 0 {
			// all jobs are completed
			return
		}
		works = append(works, work)
		entries = append(entries, entry)
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()
	results, err := w.Down.BatchDownload(ctx, works)
	if err != nil {
		log.Printf("batch download error: %s\n", err)
		return
	}
	log.Printf("| Add/Error/Complete | Name\n")
	for i, r := range results {
		if len(r.Completed) > 0 {
			entry := entries[i]
			if entry.AddCompleted(r.Completed) {
				if err := w.Repo.Save(entry); err != nil {
					log.Println("save entry error:", err)
				}
			}
		}
		log.Printf("| %4d / %4d / %4d | %s\n", r.Added, r.Failed, len(r.Completed), works[i].Name)
	}
	log.Printf("| %d polled, %d dispatched\n", allCount, len(results))
}

func (w *Worker) PollSingle(rssURL string, options map[string]string) (downloader.DownloadResult, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	work, err := poller.Poll(ctx, rssURL, options)
	if err != nil {
		return downloader.DownloadResult{}, err
	}
	results, err := w.Down.BatchDownload(ctx, []*poller.Work{work})
	if err != nil {
		return downloader.DownloadResult{}, err
	}
	if len(results) != 1 {
		return downloader.DownloadResult{}, errors.New("unexpected result count")
	}
	return results[0], nil
}
