package downloader

type DownloadResult struct {
	Added     uint32
	Failed    uint32
	Running   uint32
	Completed []string
	Removed   []string
}

func (r DownloadResult) HasUpdate() bool {
	return r.Added > 0
}

func (r DownloadResult) HasError() bool {
	return r.Failed > 0
}
