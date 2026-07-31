package download

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

func TestProgressReader_RendersAndClears(t *testing.T) {
	var out bytes.Buffer
	src := strings.NewReader(strings.Repeat("x", 1<<20))
	pr := newProgressReader(src, "tool.tar.gz", 1<<20)
	pr.out = &out
	pr.lastRender = time.Time{} // force an immediate render on first read

	if _, err := io.ReadAll(pr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "tool.tar.gz") {
		t.Fatalf("progress line must carry the asset label, got %q", out.String())
	}
	pr.done()
	if !strings.HasSuffix(out.String(), "\r\033[K") {
		t.Fatal("done() must clear the progress line")
	}
}

func TestProgressReader_UnknownLengthShowsVolume(t *testing.T) {
	var out bytes.Buffer
	pr := newProgressReader(strings.NewReader("data"), "x", -1)
	pr.out = &out
	pr.lastRender = time.Time{}
	_, _ = io.ReadAll(pr)
	if !strings.Contains(out.String(), "B") {
		t.Fatalf("unknown length must render volume, got %q", out.String())
	}
	if strings.Contains(out.String(), "%") {
		t.Fatal("unknown length must not invent a percentage")
	}
}

func TestHumanBytes(t *testing.T) {
	for in, want := range map[int64]string{
		512:     "512 B",
		2 << 10: "2.0 KB",
		3 << 20: "3.0 MB",
		2 << 30: "2.0 GB",
	} {
		if got := humanBytes(in); got != want {
			t.Errorf("humanBytes(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestAssetLabel(t *testing.T) {
	if got := assetLabel("https://github.com/x/releases/download/v1/helm-v3.16.2.tar.gz"); got != "helm-v3.16.2.tar.gz" {
		t.Fatalf("got %q", got)
	}
	if got := assetLabel("::bad::"); got != "download" {
		t.Fatalf("unparseable URL must fall back, got %q", got)
	}
}
