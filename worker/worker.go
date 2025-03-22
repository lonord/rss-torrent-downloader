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

type SubscriptionRepo interface {
	Query(fn func(id, rssURL string, options map[string]string)) error
	Save(id, rssURL string, options map[string]string) error
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
	updateCount := 0
	errorCount := 0
	allCount := 0
	works := []*poller.Work{}
	w.Repo.Query(func(id, rssURL string, options map[string]string) {
		allCount++
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		work, err := poller.Poll(ctx, rssURL, options)
		if err != nil {
			log.Printf("poll %s error: %s\n", rssURL, err)
			return
		}
		works = append(works, work)
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()
	results, err := w.Down.BatchDownload(ctx, works)
	if err != nil {
		log.Printf("batch download error: %s\n", err)
		return
	}
	for _, r := range results {
		if r.HasError() {
			errorCount++
		} else if r.HasUpdate() {
			updateCount++
		}
	}
	log.Printf("polled %d, updated %d, error %d\n", allCount, updateCount, errorCount)
}

func (w *Worker) PollSingle(rssURL string, options map[string]string, keep bool) (downloader.DownloadResult, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	work, err := poller.Poll(ctx, rssURL, options)
	if err != nil {
		return downloader.DownloadResult{}, err
	}
	if !keep {
		work.Aria2Opt["force-save"] = "false"
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
