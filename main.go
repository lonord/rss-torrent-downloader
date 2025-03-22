package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/lonord/rss-torrent-downloader/downloader"
	"github.com/lonord/rss-torrent-downloader/flagx"
	"github.com/lonord/rss-torrent-downloader/repo"
	"github.com/lonord/rss-torrent-downloader/webapi"
	"github.com/lonord/rss-torrent-downloader/worker"
)

const (
	ARIA2_SERVER = "http://127.0.0.1:6800"
)

var (
	appName    = "rss-torrent-dl"
	appVersion = "dev"
	buildTime  = ""
)

var flags struct {
	version      bool
	subscription string
	dir          string
	aria2        string
	secret       string
	interval     int
	httpAddr     string
}

func init() {
	flag.BoolVar(&flags.version, "version", false, "show version")
	flag.StringVar(&flags.subscription, "subscription", "subscription", "`directory` for reading subscription files")
	flag.StringVar(&flags.dir, "dir", "", "aria2 download directory, empty for server default")
	flag.StringVar(&flags.aria2, "aria2", ARIA2_SERVER, "`addr` for connecting video downloader server")
	flag.StringVar(&flags.secret, "secret", "", "aria2 secret token")
	flag.IntVar(&flags.interval, "interval", 60, "interval of `minutes` to poll")
	flag.StringVar(&flags.httpAddr, "http", ":6900", "`addr` for http server")
}

func main() {
	flagx.Parse(flagx.EnableFile("config"), flagx.EnableEnv("RSS_TORRENT_DL"), flagx.ExcludeFlag("version"))
	if flags.version {
		fmt.Printf("%s version %s build on %s %s/%s\n", appName, appVersion, buildTime, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	w := &worker.Worker{
		Repo:     &repo.FileRepo{Dir: flags.subscription},
		Interval: time.Minute * time.Duration(flags.interval),
		Down: &downloader.Aria2Downloader{
			URL:    flags.aria2,
			Secret: flags.secret,
			Dir:    flags.dir,
		},
	}
	httpServer := &webapi.HTTPServer{
		Addr:   flags.httpAddr,
		Worker: w,
	}
	go httpServer.Run()
	w.Run()
}
