package torrent

import (
	"io"
	"os"
	"testing"
)

func TestCalculateInfoHash1(t *testing.T) {
	f, err := os.Open("testdata/a.torrent")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	infoHash, err := calculateInfoHash(b)
	if err != nil {
		t.Fatal(err)
	}
	if infoHash != "245211c98e3f5d99cb9cf306e1133f134dbd0bcc" {
		t.Errorf("infoHash = %s; want 245211c98e3f5d99cb9cf306e1133f134dbd0bcc", infoHash)
	}
}

func TestCalculateInfoHash2(t *testing.T) {
	f, err := os.Open("testdata/b.torrent")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	infoHash, err := calculateInfoHash(b)
	if err != nil {
		t.Fatal(err)
	}
	if infoHash != "448057b2e83c50287c861697872888f75741f9ec" {
		t.Errorf("infoHash = %s; want 448057b2e83c50287c861697872888f75741f9ec", infoHash)
	}
}
