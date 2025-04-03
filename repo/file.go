package repo

import (
	"encoding/json"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/lonord/rss-torrent-downloader/worker"
)

type FileRepo struct {
	Dir string
}

func (r *FileRepo) Query(fn func(*worker.SubscriptionEntry)) error {
	return filepath.Walk(r.Dir, func(path string, info fs.FileInfo, err0 error) error {
		if !info.IsDir() && shouldIncludeExt(filepath.Ext(path)) {
			file, err := os.Open(path)
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
				entry.ID = path
				fn(&entry)
			}
		}
		return nil
	})
}

func (r *FileRepo) Save(entry *worker.SubscriptionEntry) error {
	target := entry.ID
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(entry)
}

func (r *FileRepo) Delete(path string) error {
	return os.Remove(path)
}

func shouldIncludeExt(ext string) bool {
	return ext == ".json"
}
