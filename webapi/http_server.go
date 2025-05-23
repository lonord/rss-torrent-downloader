package webapi

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/lonord/rss-torrent-downloader/worker"
)

type HTTPServer struct {
	Addr   string
	Worker *worker.Worker
}

func (s *HTTPServer) Run() {
	http.HandleFunc("/submit", s.handleSubmit)
	http.HandleFunc("/list", s.handleList)
	http.HandleFunc("/add", s.handleAdd)
	http.HandleFunc("/del", s.handleDelete)
	http.ListenAndServe(s.Addr, nil)
}

func (s *HTTPServer) handleSubmit(w http.ResponseWriter, r *http.Request) {
	handleJSON(w, func() (interface{}, error) {
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
		_, rssURL, options, err := parseURLAndOptions(r.Form)
		if err != nil {
			return nil, err
		}
		result, err := s.Worker.PollSingle(rssURL, options)
		if err != nil {
			return nil, err
		}
		log.Printf("webapi: submit success %s, %+v\n", rssURL, options)
		return result, nil
	})
}

func (s *HTTPServer) handleList(w http.ResponseWriter, r *http.Request) {
	handleJSON(w, func() (interface{}, error) {
		list := []interface{}{}
		s.Worker.Repo.Query(func(entry *worker.SubscriptionEntry) {
			item := map[string]interface{}{
				"id":        entry.ID,
				"rss":       entry.RssURL,
				"options":   entry.Options,
				"completed": len(entry.Completed),
			}
			list = append(list, item)
		})
		return map[string]interface{}{"result": list}, nil
	})
}

func (s *HTTPServer) handleAdd(w http.ResponseWriter, r *http.Request) {
	handleJSON(w, func() (interface{}, error) {
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
		id, rssURL, options, err := parseURLAndOptions(r.Form)
		if err != nil {
			return nil, err
		}
		entry := &worker.SubscriptionEntry{
			ID:      id,
			Options: options,
			RssURL:  rssURL,
		}
		if err := s.Worker.Repo.Save(entry); err != nil {
			return nil, err
		}
		if _, err := s.Worker.PollSingle(rssURL, options); err != nil {
			return nil, err
		}
		log.Printf("webapi: add success %s, %+v\n", rssURL, options)
		return map[string]string{"result": "ok"}, nil
	})
}

func (s *HTTPServer) handleDelete(w http.ResponseWriter, r *http.Request) {
	handleJSON(w, func() (interface{}, error) {
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
		id := r.FormValue("id")
		if id == "" {
			return nil, errors.New("missing id")
		}
		if err := s.Worker.Repo.Delete(id); err != nil {
			return nil, err
		}
		log.Printf("webapi: delete success %s\n", id)
		return map[string]string{"result": "ok"}, nil
	})
}

func handleJSON(w http.ResponseWriter, fn func() (interface{}, error)) {
	h := w.Header()
	h.Set("Content-Type", "application/json; charset=UTF-8")
	data, err := fn()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}
	b, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}
	w.Write(b)
}

func parseURLAndOptions(form url.Values) (string, string, map[string]string, error) {
	options := map[string]string{}
	var rssURL string
	var name string
	for k, v := range form {
		if len(v) == 0 || (len(v) == 1 && v[0] == "") {
			continue
		}
		if k == "rss" || k == "url" {
			rssURL = v[0]
			continue
		}
		if k == "name" {
			name = v[0]
			continue
		}
		options[k] = v[0]
	}
	if rssURL == "" {
		return "", "", nil, errors.New("missing rss url")
	}
	var id string
	if name != "" {
		id = name
	} else {
		hash := md5.Sum([]byte(rssURL))
		id = hex.EncodeToString(hash[:])
	}
	return id, rssURL, options, nil
}
