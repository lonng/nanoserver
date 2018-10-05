package mahjong

import (
	"fmt"
)

var tileNames = []string{"条", "筒", "万"}

type Tile struct {
	Id    int
	Suit  int //花色
	Rank  int //点数
	Index int //索引(1~9, 11~19)
}

func (t *Tile) String() string {
	return fmt.Sprintf("%d%s", t.Rank, tileNames[t.Suit])
}

func (t *Tile) Equals(other *Tile) bool {
	return t.Index == other.Index
}

func TileFromIndex(idx int) *Tile {
	if idx < 0 || idx > MaxTileIndex || idx%10 == 0 {
		return nil
	}

	return &Tile{
		Suit:  idx / 10,
		Rank:  idx % 10,
		Index: idx,
	}
}

func IndexFromID(id int) int {
	if id < 0 {
		panic(fmt.Errorf("ilegal tile id: %d", id))
	}
	var (
		tmp = id / 4
		h   = tmp / 9
		v   = tmp%9 + 1
		i   = h*10 + v
	)

	return i
}

//id: 0~3 -> 1条  4~7 -> 2条 ...
//0~35 =>条 36~71 =>筒
func TileFromID(id int) *Tile {
	if id < 0 {
		panic("illegal tile id")
	}

	var (
		tmp = id / 4
		h   = tmp / 9
		v   = tmp%9 + 1
		i   = h*10 + v
	)

	return &Tile{Suit: h, Rank: v, Index: i, Id: id}
}
