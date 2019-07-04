package mahjong

import (
	"fmt"
	"reflect"
	"testing"
)

func TestIndexes_Sort(t *testing.T) {
	var indexes = Indexes{2, 3, 87, 5, 2, 2, 2, 1, 74, 29, 39, 56, 23, 91}
	Sort()
	fmt.Printf("%+v", indexes)
}

func BenchmarkIndexes_Sort(b *testing.B) {
	var indexes = Indexes{2, 3, 87, 5, 2, 2, 2, 1, 74, 29, 39, 56, 23, 91}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Sort()
	}
}

func TestIndexes_MakeUsed(t *testing.T) {
	var indexes = Indexes{2, 3, 87, 5, 2, 2, 2, 1, 74, 29, 39, 56, 23, 91}
	Mark(5, 6, 7)
	if u := UnmarkedCount(); u != len(indexes)-3 {
		t.Fatalf("unused: %v", u)
	}
	Reset()
	if u := UnmarkedCount(); u != len(indexes) {
		t.Fatalf("unused: %v", u)
	}
	if !reflect.DeepEqual(indexes, Indexes{2, 3, 87, 5, 2, 2, 2, 1, 74, 29, 39, 56, 23, 91}) {
		t.Fatalf("not equal")
	}
}

func BenchmarkIndexes_UnusedCount(b *testing.B) {
	var indexes = Indexes{2, 3, 87, 5, 2, 2, 2, 1, 74, 29, 39, 56, 23, 91}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarkedCount()
	}
}

func TestIndexes_Unused(t *testing.T) {
	var indexes = Indexes{2, 3, 87, 5, 2, 2, 2, 1, 74, 29, 39, 56, 23, 91}
	var ret, count = UnmarkedSequence()
	if count != 3 {
		t.Fatalf("unexpect count: %d", count)
	}
	if !reflect.DeepEqual(ret, [3]byte{2, 3, 87}) {
		t.Fatalf("expece equal: %+v", ret)
	}

	Mark(1, 2)
	ret, count = UnmarkedSequence()
	if count != 3 {
		t.Fatalf("unexpect count: %d", count)
	}
	if !reflect.DeepEqual(ret, [3]byte{2, 5, 2}) {
		t.Fatalf("expece equal: %+v", ret)
	}

	Reset()
	ret, count = UnmarkedSequence()
	if count != 3 {
		t.Fatalf("unexpect count: %d", count)
	}
	if !reflect.DeepEqual(ret, [3]byte{2, 3, 87}) {
		t.Fatalf("expece equal: %+v", ret)
	}
}

func BenchmarkIndexes_Unused(b *testing.B) {
	var indexes = Indexes{2, 3, 87, 5, 2, 2, 2, 1, 74, 29, 39, 56, 23, 91}
	for i := 5; i < len(indexes); i++ {
		if i%2 == 0 {
			continue
		}
		Mark(i)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		UnmarkedSequence()
	}
}
