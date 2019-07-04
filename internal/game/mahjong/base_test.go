package mahjong

import (
	"testing"

	"github.com/lonng/nanoserver/protocol"
)

func _TestBase_CanWinBySelfDrawing(t *testing.T) {
	t.SkipNow()
	tables := []int{4, 6, 15, 15, 7, 4, 7, 6, 17, 17, 18, 18, 16, 16}

	if CanZimo(tables) != true {
		t.FailNow()
	}
}

func _TestBase_CanWinByOtherDiscard(t *testing.T) {
	t.SkipNow()

	onHand := []int{5, 15, 6, 2, 12, 16, 14, 3, 7, 3, 1, 3, 12}
	discard := 3

	if !CanHu(onHand, discard) {
		t.FailNow()
	}
}

func _TestBase_IsReadyTiles(t *testing.T) {
	t.SkipNow()
	row := []int{16, 4, 17, 4, 17, 3, 3, 17, 16, 16}

	if !IsTing(row) {
		t.FailNow()
	}
}

func _TestBase_Tings(t *testing.T) {
	//origin := []int {5, 6, 14, 8, 16, 15, 3, 15, 13, 4, 18, 17, 15, 15}

	//origin := []int {5, 6, 14, 8, 16, 3, 13, 4, 18, 17, 2}
	//5条 2条 8筒 7筒 2条 9筒 发财 6条 1条 发财 7条
	origin := []int{5, 2, 18, 17, 2, 19, 23, 6, 1, 23, 7}

	size := len(origin)

	f := func(origin []int, i int) ([]int, int) {
		size := len(origin)
		temp := make([]int, size)
		copy(temp, origin)

		idx := temp[i]

		copy(temp[i:], temp[i+1:])
		temp[len(temp)-1] = 0
		temp = temp[:len(temp)-1]
		return temp, idx

	}

	for i := 0; i < size; i++ {
		mjIndexs, idx := f(origin, i)

		huIndexs := TingTiles(mjIndexs)
		if len(huIndexs) > 0 {

			t.Logf("出牌: %d 和牌: %+v 手牌: %+v", idx, huIndexs, mjIndexs)
		} else {
			t.Log("fuck")
		}
	}
}

// #issue: 2017-08-19/916829#6
func TestIsTing(t *testing.T) {
	ctx := &Context{
		WinningID:         34,
		NewDrawingID:      16,
		NewOtherDiscardID: 34, ResultType: 0,
		Opts: &protocol.DeskOptions{
			MaxRound: 8,
			MaxFan:   3,
		},
		LastHint: &protocol.Hint{
			Ops: []protocol.Op{{Type: 4, TileIDs: []int{34}}, {Type: 5, TileIDs: []int{}}, {Type: 2, TileIDs: []int{34}}},
		},
	}

	//2条 4条 2筒 9条 5筒 1条 1筒 5筒 9条 6条 3筒 5条 3条
	onHand := Indexes{2, 4, 12, 9, 15, 1, 11, 15, 9, 6, 13, 5, 3, 9}
	Multiple(ctx, onHand, Indexes{})
}

func TestBase_IsYJ(t *testing.T) {
	tests := []struct {
		onhand, pongkong Indexes
		isYaoJiu         bool
	}{
		{
			Indexes{1, 2, 3, 1, 1, 9, 9, 9, 7, 8, 9},
			Indexes{11, 11, 11},
			true,
		},
		{
			Indexes{1, 2, 3, 1, 1, 9, 9, 9, 7, 8, 9},
			Indexes{12, 12, 12},
			false,
		},
		{
			Indexes{1, 2, 3, 1, 1, 9, 9, 9, 7, 8, 9},
			Indexes{11, 11, 11, 11},
			true,
		},
		{
			Indexes{1, 2, 3, 2, 2, 9, 9, 9, 7, 8, 9},
			Indexes{11, 11, 11},
			false,
		},
		{
			Indexes{1, 2, 3, 1, 1, 1, 9, 9, 7, 8, 9},
			Indexes{11, 11, 11},
			true,
		},
		{
			Indexes{4, 2, 3, 1, 1, 9, 9, 9, 7, 8, 9},
			Indexes{11, 11, 11},
			false,
		},
		{
			Indexes{7, 8, 9, 11, 11, 17, 17, 18, 18, 19, 19},
			Indexes{1, 1, 1},
			true,
		},
		{
			Indexes{1, 2, 3, 7, 8, 9, 11, 11, 17, 17, 18, 18, 19, 19},
			Indexes{},
			true,
		},
	}

	for _, c := range tests {
		if result := isYJ(c.onhand, c.pongkong); result != c.isYaoJiu {
			t.Fatalf("isYJ, expect=%t, got=%t", c.isYaoJiu, result)
		}
	}
}
