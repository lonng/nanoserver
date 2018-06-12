package util

import (
	"bytes"
	"testing"
)

func TestTrimSpace(t *testing.T) {
	src := "  a\ta a \t  \n  \t\r  \r\n\n  b  b\t b \t"
	expect := "a\ta ab  b\t b"
	expectBytes := []byte(expect)

	if dst := TrimSpaceString(src); dst != expect {
		t.Errorf("TrimSpaceString(%q):\nhave %q\nwant %q\n", src, dst, expect)
	}
	if dst := TrimSpace([]byte(src)); !bytes.Equal(dst, expectBytes) {
		t.Errorf("TrimSpace(%q):\nhave %q\nwant %q\n", src, dst, expectBytes)
	}
}
