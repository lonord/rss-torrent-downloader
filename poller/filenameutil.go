package poller

import "regexp"

var (
	fileNameFormatReg *regexp.Regexp
)

func init() {
	fileNameFormatReg = regexp.MustCompile("\\/|\\\\|:|\\*|\\?|\"|<|>")
}

func formatFileName(name string) string {
	n := name
	for len([]byte(n)) > 240 {
		n = n[:len(n)-1]
	}
	return fileNameFormatReg.ReplaceAllString(n, "_")
}
