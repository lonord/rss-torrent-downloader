package repo

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/lonord/rss-torrent-downloader/worker"
)

const fExt = ".json"

type FileRepo struct {
	Dir string
}

func (r *FileRepo) Query(fn func(*worker.SubscriptionEntry)) error {
	dirEntries, err := os.ReadDir(r.Dir)
	if err != nil {
		return err
	}
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() || filepath.Ext(dirEntry.Name()) != fExt {
			continue
		}
		file, err := os.Open(filepath.Join(r.Dir, dirEntry.Name()))
		if err != nil {
			log.Println("read file error:", err)
			return nil
		}
		defer file.Close()
		var entry worker.SubscriptionEntry
		if err := json.NewDecoder(file).Decode(&entry); err != nil {
			log.Println("decode file error:", err)
			return nil
		}
		if entry.RssURL != "" {
			entry.ID = strings.TrimSuffix(dirEntry.Name(), fExt)
			fn(&entry)
		}
	}
	return nil
}

func (r *FileRepo) Save(entry *worker.SubscriptionEntry) error {
	p := path.Join(r.Dir, entry.ID+fExt)
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(entry)
}

func (r *FileRepo) Delete(id string) error {
	p := path.Join(r.Dir, id+fExt)
	return os.Remove(p)
}
