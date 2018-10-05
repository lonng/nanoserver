package mahjong

import (
	"strings"
)

type Indexes []int // 麻将的index

func quickSort(values []int, left int, right int) {
	if left < right {
		temp := values[left]
		i, j := left, right
		for {
			for values[j] >= temp && i < j {
				j--
			}
			for values[i] <= temp && i < j {
				i++
			}

			if i >= j {
				break
			}

			values[i], values[j] = values[j], values[i]
		}

		values[left] = values[i]
		values[i] = temp

		quickSort(values, left, i-1)
		quickSort(values, i+1, right)
	}
}

func (indexes Indexes) Sort() {
	quickSort(indexes, 0, len(indexes)-1)
}

func (indexes Indexes) Mark(is ...int) {
	for _, i := range is {
		indexes[i] |= 0x80
	}
}

func (indexes Indexes) Unmark(is ...int) {
	for _, i := range is {
		indexes[i] &^= 0x80
	}
}

func (indexes Indexes) UnmarkedCount() int {
	var count int
	for i := 0; i < len(indexes); i++ {
		if indexes[i]&0x80 != 0 {
			continue
		}
		count++
	}
	return count
}

type IndexInfo struct {
	Index int
	I     int
}

// 返回一个刻子, 并返回数量
func (indexes Indexes) UnmarkedSequence() ([3]IndexInfo, int) {
	var count int
	var ret = [3]IndexInfo{}
	var prev int
	for i := 0; i < len(indexes); i++ {
		index := indexes[i]
		if index&0x80 != 0 {
			continue
		}
		if count < 1 || prev+1 == index {
			prev = index
			ret[count] = IndexInfo{Index: index, I: i}
			count++
		}
		if count == len(ret) {
			break
		}
	}
	return ret, count
}

func (indexes Indexes) UnmarkedTriplet() ([3]IndexInfo, int) {
	var count int
	var ret = [3]IndexInfo{}
	var prev int
	for i := 0; i < len(indexes); i++ {
		index := indexes[i]
		if index&0x80 != 0 {
			continue
		}
		if count < 1 || prev == index {
			prev = index
			ret[count] = IndexInfo{Index: index, I: i}
			count++
		}
		if count == len(ret) {
			break
		}
	}
	return ret, count
}

// 返回所有未使用的Index, 并返回数量
func (indexes Indexes) Unmarked() ([14]IndexInfo, int) {
	var count int
	var ret = [14]IndexInfo{}
	for i := 0; i < len(indexes); i++ {
		index := indexes[i]
		if index&0x80 != 0 {
			continue
		}
		ret[count] = IndexInfo{Index: index, I: i}
		count++
	}
	return ret, count
}

func (indexes Indexes) UnmarkedString() string {
	var ret []string
	for i := 0; i < len(indexes); i++ {
		index := indexes[i]
		if index&0x80 != 0 {
			continue
		}
		ret = append(ret, TileFromIndex(index).String())
	}
	return strings.Join(ret, ", ")
}

func (indexes Indexes) String() string {
	var ret []string
	for i := 0; i < len(indexes); i++ {
		index := indexes[i]
		if index&0x80 != 0 {
			index &^= 0x80
		}
		ret = append(ret, TileFromIndex(index).String())
	}
	return strings.Join(ret, ", ")
}

func (indexes Indexes) TileString(i int) string {
	index := indexes[i]
	if index&0x80 != 0 {
		index &^= 0x80
	}
	return TileFromIndex(index).String()
}

func (indexes Indexes) Reset() {
	for i := 0; i < len(indexes); i++ {
		indexes[i] &^= 0x80
	}
}
