package mahjong

import (
	"math/rand"
	"sort"
	"strings"
	"time"
)

//每种花色(条,筒)最多9种牌型,但牌是没有0点的共计 9+9
// 1-9: 条
// 11-19: 筒
// 21-29: 万
const MaxTileIndex = 29

type Mahjong []*Tile

func (m Mahjong) Len() int {
	return len(m)
}

func (m Mahjong) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m Mahjong) Less(i, j int) bool {
	return m[i].Id < m[j].Id
}

func (m Mahjong) String() string {
	res := make([]string, len(m))
	for i := range m {
		res[i] = m[i].String()
	}
	return strings.Join(res, " ")
}

func (m Mahjong) Shuffle() {
	s := rand.New(rand.NewSource(time.Now().Unix()))
	for i := range m {
		j := s.Intn(len(m))
		m[i], m[j] = m[j], m[i]
	}
}

func (m Mahjong) Sort() {
	sort.Sort(m)
}

func (m Mahjong) Indexes() []int {
	idx := make([]int, len(m))
	for i, t := range m {
		idx[i] = t.Index
	}
	return idx
}

func (m Mahjong) Ids() []int {
	ids := make([]int, len(m))
	for i, t := range m {
		ids[i] = t.Id
	}
	return ids
}

func RemoveId(m *Mahjong, tid int) {
	size := len(*m)

	i := 0
	for ; i < size; i++ {
		if (*m)[i].Id == tid {
			break
		}
	}

	if i == size {
		return
	}

	*m = append((*m)[:i], (*m)[i+1:]...)
}

//根据id索引创建
func FromID(ids []int) Mahjong {
	mj := make(Mahjong, len(ids))
	for i, idx := range ids {
		mj[i] = TileFromID(idx)
	}
	return mj
}

func init() {
	rand.Seed(time.Now().Unix())
}
