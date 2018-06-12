package basex

import (
	"strconv"
	"testing"
)

func TestBasexSuccess(t *testing.T) {
	cases := []struct {
		in string
	}{
		{"999999999999"},
		{"9007199254740992"},
		{"9007199254740989"},
		{"123456789012345678901234567890"},
		{"1234"},
	}
	for _, c := range cases {
		encode, err := Encode(c.in)
		if err != nil {
			t.Errorf("Encode error:%q", err)
		}

		decode, err := Decode(encode)
		if err != nil {
			t.Errorf("Decode error:%q", err)
		}
		t.Log(encode, decode)
		if c.in != decode {
			t.Errorf("Encode(%q) == %q, Decode %q", c.in, encode, decode)
		}
	}
}

func TestBasexFailure(t *testing.T) {
	cases := []struct {
		in string
	}{
		{"test/test/123"},
		{"https://tour.golang.org"},
		{"https://blog.golang.org"},
		{"http://golang.org/doc/#learning"},
	}
	for _, c := range cases {
		encode, _ := Encode(c.in)
		decode, _ := Decode(encode)
		t.Log(encode, decode)
		if c.in == decode {
			t.Errorf("Encode(%q) == %q, Decode %q", c.in, encode, decode)
		}
	}
}

func TestForLargeInputs(t *testing.T) {
	for i := 1000; i < 3000000; i++ {
		encode, err := Encode(strconv.Itoa(i))
		if err != nil {
			t.Errorf("Encode error:%q", err)
		}
		decode, err := Decode(encode)
		if err != nil {
			t.Errorf("Decode error:%q", err)
		}
		if strconv.Itoa(i) != decode {
			t.Errorf("Encode(%q) == %q, Decode %q", i, encode, decode)
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	s := "9007199254740989"
	for n := 0; n < b.N; n++ {
		_, _ = Encode(s)
	}
}

func BenchmarkDecode(b *testing.B) {
	s := "2aYls9bkamJJSwhr0"
	for n := 0; n < b.N; n++ {
		_, _ = Decode(s)
	}
}
