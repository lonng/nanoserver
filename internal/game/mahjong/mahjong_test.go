package mahjong

import (
	"testing"
)

func TestNew(t *testing.T) {
	for i := 0; i < 72; i++ {
		mj := TileFromID(i)
		if Suit > 1 {
			t.Fail()
		}
		if Rank > 9 {
			t.Fail()
		}
	}
}
