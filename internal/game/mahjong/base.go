package mahjong

import (
	"fmt"

	"github.com/lonng/nanoserver/protocol"
)

const (
	PingHu   = iota
	PongPong // 碰碰胡：由4副刻子（或杠）、将牌组成的胡牌
	QiDui    // 七对： 7个对子组成的的胡牌
	LongQiDui
	ShuangLongQiDui
	HaoHuaLongQiDui
	QYS          // 清一色：由一种花色的牌组成的胡牌
	HaiDiLao     // 海底捞：抓棹面上最后一张牌的人正好能胡牌
	GangShangHua // 杠上花： 杠来自己要胡的牌胡牌
)

// 番型对应的番数
var points = [...]int{
	PingHu:          0, // 平胡
	PongPong:        1, // 大对子：由4副刻子（或杠）、将牌组成的胡牌。×2。
	QiDui:           2, // 七对： 7个对子组成的的胡牌。×4
	LongQiDui:       3,
	ShuangLongQiDui: 4,
	HaoHuaLongQiDui: 5,
	QYS:             2, // 清一色：由一种花色的牌组成的胡牌。×4
	HaiDiLao:        1,
	GangShangHua:    1,
}

var descriptions = [...]string{
	QiDui:           "七对", // 七对： 7个对子组成的的胡牌。×4
	LongQiDui:       "龙七对",
	ShuangLongQiDui: "双龙七对",
	HaoHuaLongQiDui: "豪华龙七对",
}

// 返回番数
func Multiple(ctx *Context, onHand, pongKong Indexes) int {
	if len(onHand)%3 != 2 {
		panic("error tile count")
	}

	ctx.Desc = []string{}

	// 选项
	opts := ctx.Opts

	ms := NewStats(onHand, pongKong)
	// 番数
	multiple := 0

	// 检测清一色
	isQYS := isQingYiSe(ms)
	if isQYS {
		point := points[QYS]
		multiple += point
		ctx.Desc = append(ctx.Desc, "清一色")
	}

	println("===>", ctx.String(), "IsQYS", isQYS)
	if ctx.LastHint != nil {
		println("===>", "LastHint", ctx.LastHint.String())
	}

	if opts.Menqing {
		// 中张
		if isZhongzhang(ms) {
			println("===>", "中张", "+1")
			ctx.Desc = append(ctx.Desc, "中张")
			multiple++
		}
		// 门清
		if len(pongKong) == 0 {
			println("===>", "门清", "+1")
			ctx.Desc = append(ctx.Desc, "门清")
			multiple++
		}
	}

	// 杠上花/杠上炮单人只可能出现一次
	println("===>", "IsLastTile || IsGangShangHua || IsGangShangPao || IsQiangGangHu", "+1")
	if ctx.IsLastTile {
		multiple++
		ctx.Desc = append(ctx.Desc, "海底")
	}
	if ctx.IsGangShangHua {
		multiple++
		ctx.Desc = append(ctx.Desc, "杠上花")
	}
	if ctx.IsGangShangPao {
		multiple++
		ctx.Desc = append(ctx.Desc, "杠上炮")
	}
	if ctx.IsQiangGangHu {
		multiple++
		ctx.Desc = append(ctx.Desc, "抢杠胡")
	}

	gangCount := gangCount(ms)
	// 除7对外,所有的和牌型都是可以分数组合
	if isQiDui(ms) {
		point := points[QiDui+gangCount]
		multiple += int(point)
		println("===>", "七对", "+", point)
		ctx.Desc = append(ctx.Desc, descriptions[QiDui+gangCount])

		// 将七对
		if opts.Jiangdui && is258(ms) {
			println("===>", "将对", "+2")
			multiple += 2
			ctx.Desc = append(ctx.Desc, "将对")
		}
		return multiple
	}

	if isDaDui(ms) {
		point := points[PongPong]
		multiple += point
		println("===>", "碰碰胡", "+1")
		ctx.Desc = append(ctx.Desc, "碰碰胡")

		// 大对子两番选项
		if opts.Pengpeng {
			println("===>", "大对子两番", "+1")
			multiple++
		}

		// 将大对
		if opts.Jiangdui && is258(ms) {
			println("===>", "将对", "+2")
			multiple += 2
			ctx.Desc = append(ctx.Desc, "将对")
		}

		// 金钩胡
		if len(onHand) == 2 {
			println("===>", "金钩胡", "+1")
			multiple += 1
			ctx.Desc = append(ctx.Desc, "金钩胡")
		}
	} else {
		// TODO:夹心五是否可以和其他牌叠加
		if opts.Jiaxin && isJiaxin(ctx, onHand) {
			println("===>", "夹心五", "+1")
			multiple += 1
			ctx.Desc = append(ctx.Desc, "夹心五")
		}

		if opts.Yaojiu && isYJ(onHand, pongKong) {
			println("===>", "幺九", "+3")
			multiple += 3
			ctx.Desc = append(ctx.Desc, "全幺九")
		}
	}

	if gangCount > 0 {
		println("===>", "杠", "+1")
		ctx.Desc = append(ctx.Desc, fmt.Sprintf("根x%d", gangCount))
		multiple += gangCount
	}

	return multiple
}

// 返回听牌最大番数
func MaxMultiple(opts *protocol.DeskOptions, onHand, pongKong Indexes) (multiple int, index int) {
	tings := TingTiles(onHand)
	index = IllegalIndex
	multiple = -1
	for _, idx := range tings {
		handTiles := make(Indexes, len(onHand)+1)
		copy(handTiles, onHand)
		handTiles[len(onHand)] = idx

		ctx := &Context{NewOtherDiscardID: idx, Opts: opts}

		if m := Multiple(ctx, handTiles, pongKong); m > multiple {
			index = idx
			multiple = m
		}
	}

	return
}
