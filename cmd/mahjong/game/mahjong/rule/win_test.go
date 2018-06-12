package rule

import (
	"testing"

	"github.com/lonnng/nanoserver/cmd/mahjong/game/mahjong"
)

func _TestEraseValue(t *testing.T) {
	arr := []int{1, 1}

	eraseValue(arr, 0)
	if len(arr) != 2 {
		t.Failed()
	}

	eraseValue(arr, 1)
	if arr[0] != 1 {
		t.Failed()
	}

}

func TestShrink(t *testing.T) {
	table := [][]int{
		[]int{3, 3, 3},
		[]int{3, 3, 3, 4, 5},
		[]int{3, 3, 4, 5},
		[]int{3, 3, 4},
		[]int{3, 4, 5},
		//[]int{3,5,6}, //此种情况实际中不存在
		[]int{3, 4},
	}

	result := []bool{
		true,
		true,
		true,
		false,
		true,
		false,
	}

	for i, row := range table {
		if _, ret := Shrink(row); ret != result[i] {
			t.Fatalf("index: %d got: %v want: %v\n", i, ret, result[i])
		}
	}
}

func _TestGroup(t *testing.T) {
	t.SkipNow()
	table := [][]int{
		[]int{1, 2, 3, 4},
		[]int{1, 1, 1, 1},
		[]int{1, 2, 3, 4, 4, 4, 5},
		[]int{1, 3, 4, 5, 5, 7, 8, 9},
	}

	result := []int{1, 1, 1, 3}

	for i, row := range table {
		if size := len(Group(row)); size != result[i] {
			t.Fatalf("index: %d got: %v want: %v\n", i, size, result[i])
		}
	}
}

func TestIsLegal(t *testing.T) {
	table := [][]int{
		[]int{1, 2, 3},
		[]int{1, 1, 1},
		[]int{1, 1, 2},
		[]int{1, 1, 3},
		[]int{1, 1, 1, 2, 2, 2},
		[]int{1, 2, 2, 2, 2, 3},
		[]int{1, 2, 2, 2, 2, 3, 5, 5, 5},
		[]int{1, 2, 2, 2, 2, 3, 5, 6, 7},
		[]int{1, 2, 2, 2, 2, 3, 5, 6, 8},
		[]int{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7},
	}

	result := []bool{
		true,
		true,
		false,
		false,
		true,
		true,
		true,
		true,
		false,
		false,
	}

	for i, row := range table {
		if ret := IsLegal(row); ret != result[i] {
			t.Fatalf("index: %d got: %v want: %v\n", i, ret, result[i])
		}
	}
}

func _TestIsWin(t *testing.T) {
	var table = [][]int{
		//{1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4, 5, 5},
		//{1, 2, 3, 4, 5, 5, 5, 5, 6, 6, 7, 8, 9, 9},
		//{1, 1, 2, 2, 3, 3, 7, 8, 8, 8, 8, 9, 9, 9},
		//{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 5, 7, 8, 9},
		//{1, 1, 2, 2, 3, 3, 14, 14, 14, 18, 19, 6, 6, 6},
		//{1, 1, 1, 2, 2, 3, 3, 3, 4, 4, 5, 5, 5, 5},
		//{1, 1, 1, 11, 11, 11, 9, 9, 9, 19, 19, 19, 5, 5},
		//{21, 21, 21, 22, 22, 22, 23, 23, 23, 2, 3, 4, 5, 5},
		//{1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4, 5, 5},
		//{2, 2, 2, 3, 3, 3, 4, 4, 4, 5, 5},
		//{5, 5, 15, 16, 17},
		//{3,3,3,6,7,8,13,14,14,14,15,16,16,16},
		//
		//{7, 21, 1, 15, 16, 4, 4, 5, 1, 19, 4, 4, 7, 8},
		//{16, 5, 9, 6, 14, 25, 25, 15, 7, 13, 15, 9, 17, 25},
		//	{1,2,3, 4,5,6, 12,13,14, 11,12, 13, 14, 14},
		//	{1,2,3, 4,5,6, 7,8,9, 11,12, 13, 14, 14},
		//	{1,2,3, 4,5,5,5,6, 7,8,9, 11,12, 13},
		//	{1,2,3, 4,5,5,5,6, 7,8,9, 25,25, 25},
		//{1, 2, 3, 4, 5, 5, 5, 5, 6, 7, 8, 9, 25, 25},
		{4, 6, 21, 7, 4, 7, 6, 17, 17, 21, 18, 18, 16, 16},
	}

	result := []bool{
		//true,
		//true,
		//true,
		//true,
		//false,
		//true,
		//true,
		//true,
		//true,
		//true,
		//true,
		//true,
		//false,
		//true,
		//true,
		//true,
		false,
	}

	for i, row := range table {
		//sort.Ints(row)
		if ret := IsWinWithIndexes(row); ret != result[i] {
			t.Fatalf("index: %d got: %v want: %v\n", i, ret, result[i])
		}
	}

	debug = true
	onHand := mahjong.Indexes{12, 12, 12, 19}
	//println(fmt.Sprintf("%v", onHand))
	println(IsTing(onHand))

}
