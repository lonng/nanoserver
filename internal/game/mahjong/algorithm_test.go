package mahjong

import (
	"testing"
)

func TestCheckWin(t *testing.T) {
	cases := []struct {
		indexes Indexes
		result  bool
	}{
		{indexes: Indexes{1, 1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 9, 9, 9}, result: true},
		{indexes: Indexes{1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 9, 9, 9, 21}, result: false},
		{indexes: Indexes{31, 32, 33, 3, 4, 5, 6, 7, 8, 9, 9, 9, 21, 21}, result: false},
		{indexes: Indexes{1, 1, 1, 1, 2, 3, 3, 3, 22, 23, 25, 33, 33, 33}, result: false},
		{indexes: Indexes{1, 1, 1, 1, 2, 3, 3, 3, 22, 23, 24, 33, 33, 33}, result: true},
		{indexes: Indexes{1, 1, 1, 1, 2, 3, 3, 3, 22, 23, 24, 33, 34, 35}, result: false},
		{indexes: Indexes{1, 1, 2, 2, 3, 3, 3, 3, 4, 4, 5, 5, 7, 7}, result: true},
		{indexes: Indexes{1, 1, 2, 2, 3, 3, 3, 3, 4, 5, 6, 7, 8, 9}, result: true},
		{indexes: Indexes{11, 12, 12, 13, 13, 14, 3, 3, 4, 5, 6, 7, 8, 9}, result: true},
		{indexes: Indexes{22, 22, 23, 23, 23, 24, 24, 24, 25, 5, 5, 7, 8, 9}, result: true},
		{indexes: Indexes{2, 2, 3, 4, 5, 5, 5, 5, 6, 7, 7, 7, 7, 8}, result: true},
		{indexes: Indexes{1, 2, 3, 3, 3, 3, 4, 5, 6, 7, 7, 7, 7, 8}, result: true},
		{indexes: Indexes{1, 2, 2, 3, 3, 3, 3, 4, 6, 7, 7, 7, 7, 8}, result: true},
		{indexes: Indexes{1, 2, 2, 2, 3, 3, 3, 4, 4, 6, 7, 7, 7, 8}, result: true},
		{indexes: Indexes{1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4, 6, 7, 8}, result: true},
	}

	for _, c := range cases {
		if r := CheckWin(c.indexes); r != c.result {
			t.Fatalf("expect: %v, got: %v, indexes: %s", c.result, r, String())
		}
	}
}

func BenchmarkCheckWin(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	//indexes := Indexes{1, 1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 9, 9, 9}
	indexes := Indexes{2, 2, 3, 4, 5, 5, 5, 5, 6, 7, 7, 7, 7, 8}
	for i := 0; i < b.N; i++ {
		CheckWin(indexes)
	}
}
