package downloader

type DownloadResult struct {
	Added  uint32 `json:"added"`
	Failed uint32 `json:"failed"`
	Exists uint32 `json:"exists"`
}

func (r DownloadResult) HasUpdate() bool {
	return r.Added > 0
}

func (r DownloadResult) HasError() bool {
	return r.Failed > 0
}
