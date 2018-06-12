package rule

import (
	"sort"

	"github.com/lonnng/nanoserver/cmd/mahjong/game/mahjong"
	"github.com/lonnng/nanoserver/internal/protocol"
)

func Stats(indexes ...mahjong.Indexes) *mahjong.Stats {
	ts := &mahjong.Stats{}
	ts.FromIndex(indexes...)
	return ts
}

//是否是清一色
func isQingYiSe(ms *mahjong.Stats) bool {
	var flag byte

	for i, v := range ms {
		if v == 0 {
			continue
		}

		if i > 20 {
			flag |= 1 << 2
		} else if i > 10 {
			flag |= 1 << 1
		} else {
			flag |= 1
		}

	}

	return flag == 4 || flag == 2 || flag == 1
}

//7对, 返回是否是七对, 以及包含杠的个数
func isQiDui(ms *mahjong.Stats) bool {
	pairCount := 0

	for _, v := range ms {
		if v == 0 {
			continue
		}
		//七对
		if v != 2 && v != 4 {
			return false
		}

		if v == 2 {
			pairCount++
		} else if v == 4 {
			pairCount += 2
		}
	}

	return pairCount == 7
}

//大对子
func isDaDui(ms *mahjong.Stats) bool {
	counter := 0

	for _, v := range ms {
		if v == 0 {
			continue
		}

		// 有单牌不可能是大对子
		if v < 2 {
			return false
		}

		if v >= 3 {
			counter++
		}
	}

	return counter == 4
}

// 检查大对子和七对是不是只包含258
func is258(ms *mahjong.Stats) bool {
	for index, v := range ms {
		if v == 0 {
			continue
		}

		if mod := index % 10; mod != 2 || mod != 5 || mod != 8 {
			return false
		}
	}
	return true
}

// 胡牌时, 所有牌没有1和9
func isZhongzhang(ms *mahjong.Stats) bool {
	for index, v := range ms {
		if v == 0 {
			continue
		}

		if mod := index % 10; mod == 1 || mod == 9 {
			return false
		}
	}
	return true
}

// 是否是夹心五
func isJiaxin(ctx *mahjong.Context, onHand mahjong.Indexes) bool {
	index := mahjong.IndexFromID(ctx.NewDrawingID)
	if id := ctx.NewOtherDiscardID; id != protocol.OptypeIllegal && id >= 0 {
		index = mahjong.IndexFromID(ctx.NewOtherDiscardID)
	}

	//5,15
	if index%10 != 5 {
		return false
	}

	//默认胡5条
	willRemoveTiles := mahjong.Indexes{4, 5, 6}
	if index == 15 {
		willRemoveTiles = mahjong.Indexes{14, 15, 16}
	} else if index == 25 {
		willRemoveTiles = mahjong.Indexes{24, 25, 26}
	}

	//卡5星判断规则:
	//胡的牌必须是5条、5同
	//移除4,5,6 OR 14,15,16 OR 24,25,26后仍然可以和牌

	marker := func(tiles mahjong.Indexes, r int) {
		for i := 0; i < len(tiles); i++ {
			//只移除第一个
			if tiles[i] == r {
				tiles[i] = mahjong.IllegalIndex
				return
			}
		}

	}

	temp := make(mahjong.Indexes, len(onHand))

	for i := 0; i < len(onHand); i++ {
		temp[i] = onHand[i]
	}

	for _, t := range willRemoveTiles {
		marker(temp, t)
	}

	var tiles mahjong.Indexes

	for _, t := range temp {
		if t == mahjong.IllegalIndex {
			continue

		}

		tiles = append(tiles, t)
	}

	return IsWinWithIndexes(tiles)
}

func min(n byte, ns ...byte) byte {
	m := n
	for _, x := range ns {
		if x < m {
			m = x
		}
	}
	return m
}

// 判断是不是幺九
// 1. 排除1和9的刻字，如果有不是1和9的刻子就不可能是幺九
func isYJ(onHand, pongkong mahjong.Indexes) bool {
	pg := Stats(pongkong)
	for index, count := range pg {
		if count < 3 {
			continue
		}
		if m := index % 10; m != 1 && m != 9 {
			return false
		}
	}

	ms := Stats(onHand)
	//println(ms.String())

	// 清理顺子，如果有1就删除2/3，如果有9就删除7/8，如果剩下的对子是1/9则成功
	yao := []byte{1, 11, 21}
	for _, y := range yao {
		if count1 := ms[y]; count1 > 0 {
			count2 := ms[y+1]
			count3 := ms[y+2]
			c := min(count1, count2, count3)

			ms[y] -= c
			ms[y+1] -= c
			ms[y+2] -= c
		}
	}

	jiu := []byte{9, 19, 29}
	for _, j := range jiu {
		if count1 := ms[j]; count1 > 0 {
			count2 := ms[j-1]
			count3 := ms[j-2]
			c := min(count1, count2, count3)

			ms[j] -= c
			ms[j-1] -= c
			ms[j-2] -= c
		}
	}

	//println(ms.String())

	// 清理刻子
	for index, count := range ms {
		if count < 3 {
			continue
		}

		m := index % 10
		// 有不是1/9的刻子，不可能是幺九
		if m != 1 && m != 9 {
			return false
		}

		ms[index] -= 3
	}

	// 剩下的对子也只能是幺九
	for index, count := range ms {
		if count > 0 && count != 2 {
			return false
		}
		if count != 2 {
			continue
		}
		m := index % 10
		if m != 1 && m != 9 {
			return false
		}
	}

	//println(ms.String())
	return true
}

func gangCount(ms *mahjong.Stats) int {
	counter := 0
	for _, v := range ms {
		if v == 4 {
			counter++
		}
	}
	return counter
}

func CanHu(onHand mahjong.Indexes, discard int) bool {
	onHand = append(onHand, discard)
	sort.Ints(onHand)
	return CanZimo(onHand)

}

func CanZimo(onHand mahjong.Indexes) bool {
	ms := Stats(onHand)
	if ok := isQiDui(ms); ok {
		return true
	}

	return IsWinWithIndexes(onHand)
}

func IsTing(onHand mahjong.Indexes) bool {
	for i := 0; i <= mahjong.MaxTileIndex; i++ {
		if i%10 == 0 {
			continue
		}
		clone := make(mahjong.Indexes, len(onHand)+1)
		copy(clone, onHand)
		clone[len(onHand)] = i

		ms := Stats(clone)

		// 7对检测
		if ok := isQiDui(ms); ok {
			return true
		}

		if IsWinWithIndexes(clone) {
			return true
		}

	}
	return false
}

// 传入一副牌，返回所有的听牌
func TingTiles(onHand mahjong.Indexes) mahjong.Indexes {
	rts := make(mahjong.Indexes, 0)
	for i := 0; i <= mahjong.MaxTileIndex; i++ {
		if i%10 == 0 {
			continue
		}

		clone := make(mahjong.Indexes, len(onHand)+1)
		copy(clone, onHand)
		clone[len(onHand)] = i

		ms := Stats(clone)

		//所有的7对检测

		if ok := isQiDui(ms); ok {
			rts = append(rts, i)
			continue
		}

		if IsWinWithIndexes(clone) {
			rts = append(rts, i)
		}
	}

	return rts
}
