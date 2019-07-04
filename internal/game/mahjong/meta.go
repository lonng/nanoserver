package mahjong

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"

	"github.com/lonng/nanoserver/protocol"
)

const (
	MaxFan = -1 // 极品
	MeiHu  = -2 // 没胡牌
)

type Tiles []int //麻将内部表示(即72张牌的id号列表)

func (m Tiles) Shuffle() {
	s := rand.New(rand.NewSource(time.Now().Unix()))
	for i := range m {
		j := s.Intn(len(m))
		m[i], m[j] = m[j], m[i]
	}
}

func New(count int) Tiles {
	tiles := make(Tiles, count)

	for i := range tiles {
		tiles[i] = i
	}

	tiles.Shuffle()
	return tiles
}

type Stats [MaxTileIndex + 1]byte

func (ms *Stats) String() string {
	buf := &bytes.Buffer{}

	for i, count := range ms {
		if count == 0 {
			continue
		}
		fmt.Fprintf(buf, "%s:%d ", TileFromIndex(i), count)
	}

	return buf.String()
}

func (ms *Stats) From(mjs ...Mahjong) {
	for _, mj := range mjs {
		for _, t := range mj {
			ms[t.Index] = ms[t.Index] + 1
		}

	}
}

func (ms *Stats) FromIndex(tiles ...Indexes) {
	for _, tile := range tiles {
		for _, idx := range tile {
			if idx >= 0 && idx <= MaxTileIndex {
				ms[idx] = ms[idx] + 1
			}
		}
	}
}

func (ms *Stats) CountWithIndex(idx int) int {
	if idx < 0 || idx%10 == 0 || idx > MaxTileIndex {
		return IllegalIndex
	}
	fmt.Println("CountWithIndex", idx, ms)
	return int(ms[idx])
}

type ReadyTile struct {
	Index  int //和牌的index
	Points int //番数
}

func (rt *ReadyTile) String() string {
	return fmt.Sprintf("%v: %d番", TileFromIndex(rt.Index), rt.Points)
}

func (rt *ReadyTile) Equals(t *ReadyTile) bool {
	return rt.Index == t.Index && rt.Points == t.Points
}

type ScoreChangeType byte

type Context struct {
	WinningID         int      //自己要和的牌
	PrevOp            int      //上一个操作
	NewDrawingID      int      //最新上手的牌
	NewOtherDiscardID int      //上家打出的最新一张牌
	LastDiscardId     int      // 最新打过的牌
	Desc              []string // 描述

	IsLastTile bool //自己上手的最新一张牌,是否是桌面上的最后一张

	LastHint *protocol.Hint //最后一次提示

	ResultType int

	Opts   *protocol.DeskOptions
	DeskNo string
	Uid    int64

	Fan int // 番数, -1表示极品
	Que int // 定缺，0表示未定缺，1表示缺条/2缺筒/3缺万

	IsGangShangHua bool // 是不是杠上花
	IsGangShangPao bool // 是不是杠上炮
	IsQiangGangHu  bool // 是不是抢杠胡
}

func (c *Context) Reset() {
	c.WinningID = -1 //真实id从0开始
	c.NewOtherDiscardID = -1
	c.PrevOp = protocol.OptypeIllegal
	c.LastDiscardId = IllegalIndex
	c.IsLastTile = false
	c.LastHint = nil
	c.ResultType = 0
	c.Fan = 0
	c.Que = 0

	c.IsGangShangHua = false
	c.IsGangShangPao = false
	c.IsQiangGangHu = false
}

func (c *Context) String() string {
	return fmt.Sprintf("Uid=%d, DeskNo=%s, WinningID=%d, PrevOp=%d, NewDrawingID=%d, NewOtherDiscardID=%d, IsLastTile=%t, ResultType=%d, Opts=%#v, IsGangShangHua=%t, IsGangShangPao=%t, IsQiangGangHu=%t",
		c.Uid, c.DeskNo, c.WinningID, c.PrevOp, c.NewDrawingID, c.NewOtherDiscardID, c.IsLastTile,
		c.ResultType, c.Opts, c.IsGangShangHua, c.IsGangShangPao, c.IsQiangGangHu)
}

func (c *Context) SetPrevOp(op int) {
	c.PrevOp = op
}

type Result []int

func (res Result) String() string {

	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, "%v%v\t", TileFromIndex(res[0]), TileFromIndex(res[1]))

	for i, res := range res[2:] {
		if i%3 == 0 {
			fmt.Fprintf(buf, "\t")
		}
		fmt.Fprintf(buf, "%v", TileFromIndex(res))
	}

	return buf.String()
}

const (
	IllegalIndex = -1
)
