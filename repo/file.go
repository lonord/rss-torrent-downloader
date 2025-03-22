package repo

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type FileRepo struct {
	Dir string
}

func (r *FileRepo) Query(fn func(string, string, map[string]string)) error {
	return filepath.Walk(r.Dir, func(path string, info fs.FileInfo, err0 error) error {
		if !info.IsDir() && shouldIncludeExt(filepath.Ext(path)) {
			file, err := os.Open(path)
			if err != nil {
				log.Println("read file error:", err)
				return nil
			}
			scanner := bufio.NewScanner(file)
			var firstLine string
			options := map[string]string{}
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
					continue
				}
				if firstLine == "" {
					firstLine = line
				} else {
					parts := strings.Split(line, "=")
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])
						options[key] = value
					}
				}
			}
			if firstLine != "" {
				fn(path, firstLine, options)
			}
		}
		return nil
	})
}

func (r *FileRepo) Save(path string, rssURL string, options map[string]string) error {
	target := filepath.Join(r.Dir, path)
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintln(f, rssURL)
	for k, v := range options {
		fmt.Fprintf(f, "%s=%s\n", k, v)
	}
	return nil
}

func (r *FileRepo) Delete(path string) error {
	return os.Remove(filepath.Join(r.Dir, path))
}

func shouldIncludeExt(ext string) bool {
	return ext == ".txt" || ext == ".ini" || ext == ".conf"
}
