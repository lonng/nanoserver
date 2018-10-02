package rule

import (
	"github.com/lonnng/nanoserver/cmd/mahjong/game/mahjong"
	"github.com/lonnng/nanoserver/protocol"
)

var defaultRule = NewBase()

type Scorer interface {
	Multiple(ctx *mahjong.Context, onHand, pongKong mahjong.Indexes) int                               //算分数
	MaxMultiple(opts *protocol.DeskOptions, onHand, pongKong mahjong.Indexes) (mutiple int, index int) //最大番数
}

func Rule() Scorer {
	return defaultRule
}
