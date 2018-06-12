package url

import (
	"net/url"
	"testing"
)

func TestQueryEscape(t *testing.T) {
	bs := make([]byte, 0, 256)
	for i := 0; i < 256; i++ {
		if i == 0x20 {
			continue
		}
		bs = append(bs, byte(i))
	}
	dst1 := QueryEscape(string(bs))
	dst2 := url.QueryEscape(string(bs))
	if dst1 != dst2 {
		t.Errorf("QueryEscape failed\nhave %q\nwant %q", dst1, dst2)
		return
	}

	if have, want := QueryEscape(" "), "%20"; have != want {
		t.Errorf(`QueryEscape " " failed, have %q, want %q`, have, want)
		return
	}
}

func TestQueryEscapeBytes(t *testing.T) {
	bs := make([]byte, 0, 256)
	for i := 0; i < 256; i++ {
		bs = append(bs, byte(i))
	}
	dst1 := QueryEscape(string(bs))
	dst2 := QueryEscapeBytes(bs)
	if dst1 != string(dst2) {
		t.Errorf("QueryEscapeBytes mismatch with QueryEscape\nQueryEscapeBytes is\t%q\nQueryEscape is\t%q", dst2, dst1)
		return
	}
}
