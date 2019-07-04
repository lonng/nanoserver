package game

import (
	"fmt"

	"github.com/lonng/nano/session"
	"github.com/lonng/nanoserver/db"
	"github.com/lonng/nanoserver/db/model"
	"github.com/lonng/nanoserver/internal/game/mahjong"
	"github.com/lonng/nanoserver/pkg/async"
	"github.com/lonng/nanoserver/protocol"
	log "github.com/sirupsen/logrus"
)

type Loser struct {
	uid   int64
	score int
}

type Player struct {
	uid  int64  // 用户ID
	head string // 头像地址
	name string // 玩家名字
	ip   string // ip地址
	sex  int    // 性别
	coin int64  // 房卡数量

	// 玩家数据
	session *session.Session

	// 游戏相关字段
	onHand   mahjong.Mahjong
	pongKong mahjong.Mahjong
	chupai   mahjong.Mahjong
	ctx      *mahjong.Context

	chOperation chan *protocol.OpChoosed

	desk  *Desk //当前桌
	turn  int   //当前玩家在桌上的方位
	score int   //经过n局后,当前玩家余下的分值数,默认为1000

	logger *log.Entry // 日志
}

func newPlayer(s *session.Session, uid int64, name, head, ip string, sex int) *Player {
	p := &Player{
		uid:   uid,
		name:  name,
		head:  head,
		ctx:   &mahjong.Context{Uid: uid},
		ip:    ip,
		sex:   sex,
		score: 1000,

		logger: log.WithField(fieldPlayer, uid),

		chOperation: make(chan *protocol.OpChoosed, 1),
	}

	p.ctx.Reset()
	p.bindSession(s)
	p.syncCoinFromDB()

	return p
}

// 异步从数据库同步房卡
func (p *Player) syncCoinFromDB() {
	async.Run(func() {
		u, err := db.QueryUser(p.uid)
		if err != nil {
			p.logger.Errorf("玩家同步房卡错误, Error=%v", err)
			return
		}

		p.coin = u.Coin
		if s := p.session; s != nil {
			s.Push("onCoinChange", &protocol.CoinChangeInformation{p.coin})
		}
	})
}

// 异步扣除玩家房卡
func (p *Player) loseCoin(count int64, consume *model.CardConsume) {
	async.Run(func() {
		u, err := db.QueryUser(p.uid)
		if err != nil {
			p.logger.Errorf("扣除房卡，查询玩家错误, Error=%v", err)
			return
		}

		// 即使数据库不成功，玩家房卡数量依然扣除
		p.coin -= count
		u.Coin = p.coin
		if err := db.UpdateUser(u); err != nil {
			p.logger.Errorf("扣除房卡，更新房卡数量错误, Error=%v", err)
			return
		}

		if u.Coin != p.coin {
			p.logger.Errorf("玩家扣除房卡，同步到数据库后，发现房卡数量不一致，玩家数量=%d，数据库数量=%d", p.coin, u.Coin)
		}

		if err := db.Insert(consume); err != nil {
			p.logger.Errorf("新增消费数据错误，Error=%v Payload=%+v", err, consume)
		}

		if s := p.session; s != nil {
			s.Push("onCoinChange", &protocol.CoinChangeInformation{p.coin})
		}
	})
}

func (p *Player) setDesk(d *Desk, turn int) {
	if d == nil {
		p.logger.Error("桌号为空")
		return
	}

	p.desk = d
	p.turn = turn

	p.logger = log.WithFields(log.Fields{fieldDesk: p.desk.roomNo, fieldPlayer: p.uid})

	//全、半频道
	p.ctx.Opts = d.opts
	p.ctx.DeskNo = string(d.roomNo)

}

func (p *Player) setIp(ip string) {
	p.ip = ip
}

func (p *Player) bindSession(s *session.Session) {
	p.session = s
	p.session.Set(kCurPlayer, p)
}

func (p *Player) removeSession() {
	p.session.Remove(kCurPlayer)
	p.session = nil
}

func (p *Player) Uid() int64 {
	return p.uid
}

func (p *Player) duanPai(ids mahjong.Tiles) {
	p.onHand = mahjong.FromID(ids)
	p.logger.Debugf("游戏开局, 手牌数量=%d 手牌: %v", len(p.handTiles()), p.handTiles())
	if len(p.onHand) == 14 {
		p.ctx.NewDrawingID = p.onHand[13].Id
	}
}

func (p *Player) chuPai() int {
	var tid int

ctrl:
	p.hint([]protocol.Op{{Type: protocol.OptypeChu}}, p.tingTiles())
	select {
	case op, ok := <-p.chOperation:
		if !ok {
			return deskDissolved
		}

		if op.Type != protocol.OptypeChu {
			p.logger.Errorf("玩家操作异常，期待操作出牌，获取操作=%+v", op)
			goto ctrl
		}

		tid = op.TileID
		if tid < 0 {
			p.logger.Debugf("玩家读取到一个非法的麻将ID: ID=%+v", op)
		}

	case <-p.desk.die:
		return deskDissolved
	}

	// 删掉已经出过的牌
	for j, mj := range p.onHand {
		if mj.Id == tid {
			rest := make([]*mahjong.Tile, len(p.onHand)-1)
			copy(rest[:j], p.onHand[:j])
			copy(rest[j:], p.onHand[j+1:])
			p.onHand = rest
			break
		}
	}

	p.logger.Debugf("玩家出牌: 麻将=%v(%d) 新上手=%v 余牌=%v",
		mahjong.TileFromID(tid),
		tid,
		p.ctx.NewDrawingID,
		p.handTiles())

	p.action(protocol.OptypeChu, []int{tid})
	return tid
}

// 让玩家选择胡牌
// @param: hasHint 是否已经提示过玩家
func (p *Player) hu(tileID int, hasHint bool) int {
	// 真实玩家
	if !hasHint {
		p.hint([]protocol.Op{
			{Type: protocol.OptypeHu, TileIDs: []int{tileID}},
			{Type: protocol.OptypePass},
		})
	}
	select {
	case op, ok := <-p.chOperation:
		if !ok {
			return deskDissolved
		}

		p.ctx.SetPrevOp(op.Type)
		return op.Type

	case <-p.desk.die:
		return deskDissolved
	}

	p.logger.Debugf("玩家胡牌: 麻将=%d", tileID)
	return protocol.OptypeHu
}

// 让玩家选择碰杠
// @param: hasHint 之前是否已经发送了碰杠牌的提示, 这种情况出现在玩家同时可以碰杠胡的情况
func (p *Player) pengOrGang(tileID int, ops []protocol.Op, hasHint bool) (isPeng, isGang, isDissolve bool) {
	isPeng = false
	isGang = false
	isDissolve = false

	var (
		opType = p.ctx.PrevOp
		mjs    mahjong.Tiles
	)

	// 之前没有发送了提示
	if !hasHint {
		hints := []protocol.Op{{Type: protocol.OptypePass}}
		for _, op := range ops {
			opInfo := protocol.Op{Type: op.Type, TileIDs: []int{tileID}}
			hints = append(hints, opInfo)
		}
		//碰、杠、过
		p.hint(hints)

		select {
		case op, ok := <-p.chOperation:
			if !ok {
				isDissolve = true
				return
			}

			tileID = op.TileID
			opType = op.Type

		case <-p.desk.die:
			isDissolve = true
			return
		}
	}

	switch opType {
	case protocol.OptypeGang:
		mjs = p.brotherTiles(tileID, 4)
		p.gang(tileID)
		isGang = true
	case protocol.OptypePeng:
		mjs = p.brotherTiles(tileID, 3)
		p.peng(tileID)
		isPeng = true
	case protocol.OptypePass:
		p.logger.Debugf("玩家选择过, 麻将: %d", tileID)
		return
	}

	p.action(opType, mjs)
	return isPeng, isGang, isDissolve
}

func (p *Player) moPai() {
	id := p.desk.nextTile().Id
	mo := &protocol.MoPai{
		AccountID: p.Uid(),
		TileIDs:   []int{id},
	}

	//此时将此牌上手到ctx中,待做了明牌处理后再正式放入
	p.ctx.NewDrawingID = id

	record := &protocol.OpTypeDo{
		OpType:  protocol.OptyMoPai,
		Uid:     []int64{p.Uid()},
		TileIDs: []int{id},
	}

	// 保存快照
	p.desk.snapshot.PushAction(record)

	// 这里如果返回错误, 可能是有玩家的socket关掉了, 玩家掉线, 可能需要把玩家标记为托管
	if err := p.desk.group.Broadcast("onMoPai", mo); err != nil {
		log.Error(err)
	}

	// TODO: 确认海底捞是否是最后一个摸牌的, 其他人摸最后一张牌, 点炮是否算海底捞
	p.ctx.IsLastTile = p.desk.noMoreTile()

	//p.onHand = append(p.onHand, mahjong.TileFromID(id))
	p.logger.Debugf("玩家摸牌: 手牌=%+v 新上手=%d", p.handTiles(), p.ctx.NewDrawingID)
}

func (p *Player) handTiles() mahjong.Mahjong {
	return p.onHand
}

func (p *Player) chuTiles() mahjong.Mahjong {
	return p.chupai
}

// 检查是否有叫
func (p *Player) isTing() bool {
	return mahjong.IsTing(p.handTiles().Indexes())
}

func (p *Player) tileIDWithIndex(index int) int {
	sps := p.handTiles()
	for _, sp := range sps {
		if sp.Index == index {
			return sp.Id
		}
	}
	return -1
}

func (p *Player) allTileIDWithIndex(index int) mahjong.Tiles {
	mjs := p.handTiles()
	ids := mahjong.Tiles{}
	for _, sp := range mjs {
		if sp.Index == index {
			ids = append(ids, sp.Id)
		}
	}
	return ids
}

// 碰杠牌
func (p *Player) pgTiles() mahjong.Mahjong {
	return p.pongKong
}

func (p *Player) canWin() bool {
	newTile := mahjong.TileFromID(p.ctx.NewDrawingID)
	que := p.ctx.Que
	// 打缺的牌不能胡
	if que == newTile.Suit+1 {
		return false
	}

	// 还没有打缺不能胡
	for _, t := range p.onHand {
		if que == t.Suit+1 {
			return false
		}
	}

	canWin := mahjong.CheckWin(p.handTiles().Indexes())

	p.logger.Infof("玩家计算是否可以胡牌: 手牌=%+v, 新上手=%v, 是否可以胡=%t",
		p.handTiles(), newTile, canWin)

	return canWin
}

func (p *Player) allGang() []protocol.Op {
	tileGroup := map[int][]*mahjong.Tile{}
	ops := []protocol.Op{}

	handTiles := p.handTiles()
	pgTiles := p.pgTiles()

	for _, t := range handTiles {
		tileGroup[t.Index] = append(tileGroup[t.Index], t)
	}

	p.logger.Debugf("计算杠牌开始: 手牌=%+v 碰杠=%+v 分组=%+v", handTiles, pgTiles, tileGroup)

	//可能有多个杠而玩家选择不杠,一直提示(eg: 双龙7对)
	//暗杠
	for _, grp := range tileGroup {
		// 不为4张的排除
		if len(grp) != 4 {
			continue
		}

		//对于杠来说,len(op.TileIDs) == 0
		op := protocol.Op{Type: protocol.OptypeGang} //暗杠
		op.TileIDs = append(op.TileIDs, grp[0].Id)   //将当前id填入最终列表
		p.logger.Debugf("玩家可以杠的牌: OP=%+v", op)
		ops = append(ops, op)
	}

	pgTileGroup := map[int][]*mahjong.Tile{}
	for _, pg := range pgTiles {
		pgTileGroup[pg.Index] = append(pgTileGroup[pg.Index], pg)
	}

	//lastIndex := mahjong.IndexFromID(p.ctx.NewDrawingID)
	//巴杠
	for _, sp := range handTiles {
		//之前能杠, 但是选择碰的牌不提示杠
		//if sp.Index != lastIndex {
		//	continue
		//}

		//只将那些非忽略杠牌才进行巴杠
		if len(pgTileGroup[sp.Index]) == 3 {
			op := protocol.Op{
				Type:    protocol.OptypeGang,
				TileIDs: []int{sp.Id}, //巴杠的麻将id
			}
			ops = append(ops, op)
		}
	}

	p.logger.Debugf("计算杠牌结束: 操作=%+v 新上手=%d", ops, p.ctx.NewDrawingID)
	return ops
}

func (p *Player) gangBySelf(tid int, isXiaYu bool, loser []int64) {
	// TODO: 处理及时雨
	p.desk.scoreChangeForGang(p, loser, tid, isXiaYu)
	p.gang(tid)
}

// TODO:获取听牌
func (p *Player) tingTiles() protocol.Tings {
	handTiles := p.handTiles().Indexes()
	distinct := map[int]struct{}{}
	tings := protocol.Tings{}

	for _, index := range handTiles {
		distinct[index] = struct{}{}
	}

	//log.Debugf("手牌：%+v，去重结果：%+v", handTiles, distinct)

	// 只移除一张
	exclude := func(index int) mahjong.Indexes {
		pivot := 0
		for i := range handTiles {
			if handTiles[i] == index {
				pivot = i
				break
			}
		}

		ret := make(mahjong.Indexes, len(handTiles)-1)
		copy(ret[:pivot], handTiles[:pivot])
		copy(ret[pivot:], handTiles[pivot+1:])

		return ret
	}

	for index := range distinct {
		rest := exclude(index)
		//log.Debugf("去除：%v，剩余：%+v", index, rest)
		if ting := mahjong.TingTiles(rest); len(ting) > 0 {
			tings = append(tings, protocol.Ting{Index: index, Hu: ting})
		}
	}

	return tings
}

// 检查手牌, 是否可以杠/胡
func (p *Player) doCheckHandTiles(isNewRound bool) (int, int) {
	var (
		canGang bool
		canWin  bool
		gangIDs []int
	)

	//todo: 和杠分开？
	fn := func(pp *Player, op *protocol.OpChoosed, mjs []int) (int, int) {
		pp.action(op.Type, mjs)

		// 客户端显示胡\杠牌流水
		loseUids := pp.desk.allLosers(pp)
		if op.Type == protocol.OptypeGang {

			//此处只可能是暗杠
			return protocol.OptypeGang, op.TileID

		} else {
			pp.ctx.WinningID = pp.ctx.NewDrawingID
			//杠上花
			if pp.ctx.PrevOp == protocol.OptypeGang {
				fmt.Println("杠上花", pp.desk.roomNo, pp.Uid())
				pp.ctx.IsGangShangHua = true
			}

			//赢的基础分值,如果对手已经明牌则翻倍处理
			score := pp.scoring()

			// 自摸加底和自摸加番
			if pp.desk.opts.Zimo == "fan" {
				score *= 2
				pp.ctx.Fan++
			}

			p.logger.Infof("玩家杠胡: 分数=%d prevOp=%d", score, pp.ctx.PrevOp)

			var losers []Loser
			for _, uid := range loseUids {
				losers = append(losers, Loser{uid: uid, score: score})
			}

			pp.ctx.PrevOp = op.Type
			pp.desk.scoreChangeForHu(pp, losers, pp.ctx.WinningID, protocol.HuTypeZiMo)

			pp.ctx.ResultType = ResultZiMo
		}
		return op.Type, pp.ctx.NewDrawingID
	}

	if !isNewRound {
		//必须先判定,再放入摸的牌
		p.onHand = append(p.onHand, mahjong.TileFromID(p.ctx.NewDrawingID))
	}

	canWin = p.canWin()

	//机器人如果能和能杠,则直接和,而玩家则给他两个选择
	ops := []protocol.Op{}

	//最后一张牌,不可杠
	if !p.desk.noMoreTile() {
		// 检查暗杠和刮风
		ops = p.allGang()
	}

	//如果在明牌情况下,玩家对某(N)张可杠的牌选择了过,则以后不再提示
	canGang = len(ops) != 0
	for _, op := range ops {
		gangIDs = append(gangIDs, op.TileIDs[0])
	}

	//非杠、非和、又非明
	if !canGang && !canWin {
		return protocol.OptypePass, p.ctx.NewDrawingID
	}

	// 真实玩家处理, 刮风, 下雨, 自扣
	if canWin {
		ops = append(ops, protocol.Op{Type: protocol.OptypeHu, TileIDs: []int{p.ctx.NewDrawingID}})
	}

	//玩家可以选择:杠、胡、过、明
	ops = append(ops, protocol.Op{Type: protocol.OptypePass})

	p.hint(ops)

	select {
	case op, ok := <-p.chOperation:
		if !ok {
			return protocol.OptypePass, deskDissolved
		}

		var mjs mahjong.Tiles
		switch op.Type {
		case protocol.OptypePass:
			return op.Type, p.ctx.NewDrawingID

		case protocol.OptypeGang:

			//刮风的牌,是onHand里只可能有一张,而其他3张在"pongKong"里
			mjs = p.allTileIDWithIndex(mahjong.TileFromID(op.TileID).Index)
			//巴杠,抢杠
			if len(mjs) == 1 {
				// fixed: 巴杠也需要通知客户端
				p.action(op.Type, mjs)
				return protocol.OptypeBaGang, op.TileID //可能杠的并不是最后摸的那一张牌，即n把后才杠
			}

		case protocol.OptypeHu:
			mjs = []int{op.TileID}
		}

		return fn(p, op, mjs)

	case <-p.desk.die:
		return protocol.OptypePass, deskDissolved
	}

	return protocol.OptypePass, p.ctx.NewDrawingID
}

// 计算番数
func (p *Player) scoring() int {
	m := mahjong.Multiple(p.ctx, p.handTiles().Indexes(), p.pgTiles().Indexes())
	p.ctx.Fan = m
	return 1 << uint(m)
}

func (p *Player) maxTingScore() (int, int) {
	m, idx := mahjong.MaxMultiple(p.desk.opts, p.handTiles().Indexes(), p.pgTiles().Indexes())
	p.ctx.Fan = m
	return 1 << uint(m), idx
}

// 检查吃牌
// 是否碰(pong)/杠(kong)/胡(win)其他人的牌
func (p *Player) checkChi(tid int, chuPlayer *Player) []protocol.Op {
	tile := mahjong.TileFromID(tid)
	ret := []protocol.Op{}

	// 不能碰杠缺的牌
	if tile.Suit+1 == p.ctx.Que {
		return ret
	}

	// 检查胡牌
	if p.checkHu(tid, chuPlayer.ctx.PrevOp == protocol.OptypeGang) {
		ret = append(ret, protocol.Op{Type: protocol.OptypeHu, TileIDs: []int{tile.Id}})
	}

	sameTiles := mahjong.Mahjong{}
	tiles := p.handTiles()
	for _, sp := range tiles {
		if sp.Index == tile.Index {
			sameTiles = append(sameTiles, sp)
		}
	}

	p.logger.Debugf("其他人打牌, 检查玩家是否可以杠, 手牌=%+v 检查是否能吃的牌=%s, 相同麻将=%+v",
		tiles, tile.String(), sameTiles)
	// 明牌玩家直接检查扣牌是否可以杠

	//检查碰、杠
	if len(sameTiles) == 3 && !p.desk.noMoreTile() {
		ret = append(ret, protocol.Op{Type: protocol.OptypeGang, TileIDs: mahjong.Tiles{sameTiles[0].Id}})
	}
	if len(sameTiles) == 2 {
		ret = append(ret, protocol.Op{Type: protocol.OptypePeng, TileIDs: mahjong.Tiles{sameTiles[0].Id}})
	}

	p.logger.Debugf("计算杠牌完毕, 所有可用操作: %+v", ret)
	return ret
}

// 是否胡牌, plus表示是不是额外有番(比如抢杠，别人杠上炮)
func (p *Player) checkHu(tid int, plus bool) bool {
	index := mahjong.IndexFromID(tid)
	tiles := p.handTiles()
	que := p.ctx.Que

	// 不能胡缺牌
	tile := mahjong.TileFromID(tid)
	if tile.Suit+1 == que {
		return false
	}
	// 还没有打缺不能胡
	for _, t := range tiles {
		if que == t.Suit+1 {
			return false
		}
	}

	// 检查胡牌
	canHu := mahjong.CanHu(tiles.Indexes(), index)
	if !canHu {
		return false
	}

	// 如果可以平胡
	if p.desk.opts.Pinghu {
		return true
	}

	// 如果不能点炮平胡
	onHand := append(p.handTiles().Indexes(), index)
	old := p.ctx.NewOtherDiscardID
	p.ctx.NewOtherDiscardID = tid
	m := mahjong.Multiple(p.ctx, onHand, p.pgTiles().Indexes())
	p.ctx.NewOtherDiscardID = old

	// 有番才能胡
	return m > 0 || plus
}

func (p *Player) brotherTiles(id, count int) mahjong.Tiles {
	tile := mahjong.TileFromID(id)

	ids := mahjong.Tiles{id}
	for _, m := range p.onHand {
		if m.Index == tile.Index && len(ids) <= count {
			ids = append(ids, m.Id)
		}
	}
	return ids
}

// 杠牌
/*
1. 共有三种情况:自己摸4张，自摸3张被有点杠,自己已经碰牌后再摸最后一张
*/
func (p *Player) gang(id int) {
	tile := mahjong.TileFromID(id)
	rest := []*mahjong.Tile{}
	p.logger.Debugf("玩家杠牌, 麻将=%v 手牌=%+v 碰杠=%+v", tile, p.handTiles(), p.pgTiles())
	counter := 0
	for _, m := range p.onHand {
		if m.Index == tile.Index {
			counter++
			p.pongKong = append(p.pongKong, m)
			continue
		}
		rest = append(rest, m)
	}

	// 只有被人点杠时才需要加此牌
	if counter == 3 {
		p.pongKong = append(p.pongKong, tile)
	}
	p.onHand = rest

	p.ctx.SetPrevOp(protocol.OptypeGang)
}

// 碰牌
func (p *Player) peng(id int) {
	tile := mahjong.TileFromID(id)
	rest := []*mahjong.Tile{}
	counter := 0
	for _, m := range p.onHand {
		if m.Index == tile.Index && counter < 2 {
			counter++
			p.pongKong = append(p.pongKong, m)
			continue
		}
		rest = append(rest, m)
	}
	p.pongKong = append(p.pongKong, tile)
	p.onHand = rest

	p.logger.Debugf("玩家碰牌, 麻将=%v 手牌=%+v 碰杠=%+v", tile, p.handTiles(), p.pgTiles())

	p.ctx.SetPrevOp(protocol.OptypePeng)
}

// 玩家操作
func (p *Player) action(opType int, tiles mahjong.Tiles) {
	p.logger.Debugf("玩家选择: OpType=%d Tiles=%+v", opType, tiles)
	do := &protocol.OpTypeDo{
		Uid:     []int64{p.Uid()},
		OpType:  opType,
		TileIDs: tiles,
	}
	if err := p.desk.group.Broadcast(protocol.RouteTypeDo, do); err != nil {
		log.Error(err)
	}

	//添加操作记录
	p.desk.snapshot.PushAction(do)
}

// 提示玩家选择碰/杠/胡
func (p *Player) hint(ops []protocol.Op, args ...protocol.Tings) {
	tings := protocol.Tings{}

	// 听牌过滤
	if len(args) > 0 {
		if tings = args[0]; len(tings) > 0 {
			var temp = map[int]protocol.Ting{}
			for i := range tings {
				t := tings[i]
				temp[t.Index] = t
			}

			var filtered protocol.Tings
			for _, t := range temp {
				filtered = append(filtered, t)
			}

			tings = filtered
		}
	}

	hint := &protocol.Hint{Uid: p.Uid(), Ops: ops, Tings: tings}

	p.ctx.LastHint = hint
	p.desk.lastHintUid = p.Uid()

	if p.session == nil {
		p.logger.Warnf("玩家网络已经断开，不能通知出牌")
		return
	}

	p.logger.Debugf("玩家最后提示: Hint=%+v", hint)
	p.session.Push(protocol.RouteOpTypeHint, hint)
}

// 定缺
func (p *Player) selectDefaultQue() int {
	stats := map[int]int{}
	for _, t := range p.onHand {
		stats[t.Suit]++
	}
	p.logger.Debugf("玩家选择定缺，统计数据=%+v", stats)
	q, c := 0, stats[0]
	for suit, count := range stats {
		if count < c {
			q = suit
		}
	}
	return q + 1
}

func (p *Player) reset() {
	p.onHand = mahjong.Mahjong{}
	p.pongKong = mahjong.Mahjong{}
	p.chupai = mahjong.Mahjong{}

	// 重置channel
	close(p.chOperation)
	p.chOperation = make(chan *protocol.OpChoosed, 1)
	p.ctx.Reset()
}

// 断线重连后，同步牌桌数据
// TODO: 断线重连，已和牌玩家显示不正常
func (p *Player) syncDeskData() error {
	desk := p.desk
	data := &protocol.SyncDesk{
		Status:    desk.status(),
		Players:   []protocol.DeskPlayerData{},
		ScoreInfo: []protocol.ScoreInfo{},
	}

	markerUid := int64(0)
	lastMoPaiUid := int64(0)
	for i, player := range desk.players {
		uid := player.Uid()
		if i == desk.bankerTurn {
			markerUid = uid
		}
		if i == desk.curTurn {
			lastMoPaiUid = uid
		}
		// 有可能已经有玩家和牌
		stats := desk.roundOverTilesForPlayer(player)
		playerData := protocol.DeskPlayerData{
			Uid:        uid,
			HandTiles:  stats.Tiles,
			PGTiles:    player.pgTiles().Ids(),
			ChuTiles:   player.chuTiles().Ids(),
			LatestTile: player.ctx.NewDrawingID,
			HuPai:      player.ctx.WinningID,
			HuType:     player.ctx.ResultType,
			IsHu:       desk.wonPlayers[uid],
			Que:        player.ctx.Que,
			Score:      player.score,
		}

		// 如果自己断线重连，并且在定缺中，则发回提示，使用负数表示定缺建议选项
		if p.Uid() == uid && p.ctx.Que < 1 {
			playerData.Que = -player.selectDefaultQue()
		}

		data.Players = append(data.Players, playerData)

		score := protocol.ScoreInfo{
			Uid:   uid,
			Score: player.score,
		}

		data.ScoreInfo = append(data.ScoreInfo, score)
	}
	data.MarkerUid = markerUid
	data.LastMoPaiUid = lastMoPaiUid
	data.RestCount = desk.remainTileCount()
	data.Dice1 = desk.dice.dice1
	data.Dice2 = desk.dice.dice2
	syncUid := p.Uid()
	if lastMoPaiUid == syncUid || p.desk.lastHintUid == syncUid {
		data.Hint = p.ctx.LastHint
	}
	data.LastTileId = p.desk.lastTileId
	data.LastChuPaiUid = p.desk.lastChuPaiUid
	p.logger.Debugf("同步房间数据: %+v", data)
	return p.session.Push("onSyncDesk", data)
}
