package game

type prepareContext struct {
	sortedStatus map[int64]bool //是否已经齐牌完毕
	readyStatus  map[int64]bool //是否已经ready完毕
}

func newPrepareContext() *prepareContext {
	return &prepareContext{
		sortedStatus: map[int64]bool{},
		readyStatus:  map[int64]bool{},
	}
}

func (p *prepareContext) isReady(uid int64) bool {
	return p.readyStatus[uid]
}

func (p *prepareContext) ready(uid int64) {
	p.readyStatus[uid] = true
}

func (p *prepareContext) sorted(uid int64) {
	p.sortedStatus[uid] = true
}

func (p *prepareContext) isSorted(uid int64) bool {
	return p.sortedStatus[uid]
}

func (p *prepareContext) reset() {
	p.sortedStatus = map[int64]bool{}
	p.readyStatus = map[int64]bool{}
}
