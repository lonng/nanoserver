package mahjong

func isLegal(indexes Indexes) bool {
	triplet, tripletCount := indexes.UnmarkedTriplet() // 刻子
	if tripletCount == 3 {
		indexes.Mark(int(triplet[0].I), int(triplet[1].I), int(triplet[2].I))
		return isLegal(indexes)
	}

	sequence, sequenceCount := indexes.UnmarkedSequence() // 顺子
	if sequenceCount == 3 {
		if sequence[0].Index > 30 {
			return false // 字牌不能组合成顺子
		}
		indexes.Mark(int(sequence[0].I), int(sequence[1].I), int(sequence[2].I))
		return isLegal(indexes)
	}
	return sequenceCount == 0 && tripletCount == 0
}

func CheckWin(indexes Indexes) bool {
	indexes.Sort()
	stats := Stats{}
	stats.FromIndex(indexes)
	if isQiDui(&stats) {
		return true
	}

	var prevIndex int
	for i := 0; i < len(indexes)-1; i++ {
		if indexes[i] != indexes[i+1] || indexes[i] == prevIndex {
			continue
		}

		prevIndex = indexes[i]
		indexes.Mark(i, i+1)
		if isLegal(indexes) {
			indexes.Reset()
			return true
		}
		indexes.Reset()
	}
	return false
}
