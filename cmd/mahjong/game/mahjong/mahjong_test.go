package mahjong

import (
	"testing"
)

func TestNew(t *testing.T) {
	for i := 0; i < 72; i++ {
		mj := TileFromID(i)
		if mj.Suit > 1 {
			t.Fail()
		}
		if mj.Rank > 9 {
			t.Fail()
		}
	}
}
