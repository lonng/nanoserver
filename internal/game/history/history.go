package history

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lonng/nanoserver/db"
	"github.com/lonng/nanoserver/db/model"
	"github.com/lonng/nanoserver/protocol"
)

type SnapShot struct {
	Enter     *protocol.PlayerEnterDesk `json:"enter"`
	BasicInfo *protocol.DeskBasicInfo   `json:"basicInfo"`
	DuanPai   *protocol.DuanPai         `json:"duanPai"`
	End       *protocol.RoundOverStats  `json:"end"`

	// 如果在此遇到了gang操作就去GangScoreChanges中按序拿数据,
	// 如果遇到了hu就去HuScoreChanges中拿数据,
	// 即杠与和数据分开放
	Do []*protocol.OpTypeDo `json:"do"`

	GangScoreChanges []*protocol.GangPaiScoreChange `json:"gangScoreChanges"`
	HuScoreChanges   []*protocol.HuInfo             `json:"huScoreChanges"`
}

type History struct {
	mode         int
	beginAt      int64
	endAt        int64
	deskID       int64
	playerName0  string
	playerName1  string
	playerName2  string
	playerName3  string
	scoreChange0 int
	scoreChange1 int
	scoreChange2 int
	scoreChange3 int

	SnapShot
}

func New(deskID int64, mode int, name0, name1, name2, name3 string, basic *protocol.DeskBasicInfo, enter *protocol.PlayerEnterDesk, duan *protocol.DuanPai) *History {
	return &History{
		beginAt:     time.Now().Unix(),
		deskID:      deskID,
		mode:        mode,
		playerName0: name0,
		playerName1: name1,
		playerName2: name2,
		playerName3: name3,
		SnapShot: SnapShot{
			BasicInfo: basic,
			DuanPai:   duan,
			Enter:     enter,
		},
	}
}

func (h *History) PushAction(op *protocol.OpTypeDo) {
	h.Do = append(h.Do, op)
}

func (h *History) PushGangScoreChange(g *protocol.GangPaiScoreChange) error {
	h.GangScoreChanges = append(h.GangScoreChanges, g)
	return nil
}

func (h *History) PushHuScoreChange(hsc *protocol.HuInfo) error {
	h.HuScoreChanges = append(h.HuScoreChanges, hsc)
	return nil
}

func (h *History) SetEndStats(ge *protocol.RoundOverStats) error {
	h.End = ge

	return nil
}

func (h *History) SetScoreChangeForTurn(turn uint8, sc int) error {
	switch turn {
	case 0:
		h.scoreChange0 = sc
	case 1:
		h.scoreChange1 = sc
	case 2:
		h.scoreChange2 = sc
	case 3:
		h.scoreChange3 = sc
	default:
		return nil

	}
	return nil
}

func (h *History) Save() error {
	data, err := json.Marshal(&h.SnapShot)
	if err != nil {
		return err
	}

	t := &model.History{
		DeskId:       h.deskID,
		BeginAt:      h.beginAt,
		Mode:         h.mode,
		EndAt:        time.Now().Unix(),
		PlayerName0:  h.playerName0,
		PlayerName1:  h.playerName1,
		PlayerName2:  h.playerName2,
		PlayerName3:  h.playerName3,
		ScoreChange0: h.scoreChange0,
		ScoreChange1: h.scoreChange1,
		ScoreChange2: h.scoreChange2,
		ScoreChange3: h.scoreChange3,
		Snapshot:     string(data),
	}

	return db.InsertHistory(t)
}

type Record struct {
	ZiMoNum     int `json:"ziMo"`
	HuNum       int `json:"hu"`
	PaoNum      int `json:"pao"`
	MingGangNum int `json:"mingGang"`
	AnGangNum   int `json:"anGang"`
	TotalScore  int `json:"totalScore"`
}

//场统计
type MatchStats map[int64][]*Record

//局统计
type RoundStats map[int64]*Record

func (ps MatchStats) Push(rs RoundStats) error {
	if len(rs) == 0 {
		return nil
	}
	for uid, r := range rs {
		ps[uid] = append(ps[uid], r)
	}

	return nil
}

func (ps MatchStats) Result() map[int64]*Record {
	ret := make(map[int64]*Record)

	for p, records := range ps {
		if _, ok := ret[p]; !ok {
			ret[p] = &Record{}
		}

		for _, m := range records {

			ret[p].AnGangNum += m.AnGangNum
			ret[p].MingGangNum += m.MingGangNum
			ret[p].TotalScore += m.TotalScore
			ret[p].ZiMoNum += m.ZiMoNum
			ret[p].HuNum += m.HuNum
			ret[p].PaoNum += m.PaoNum
		}

		fmt.Printf("统计 玩家: %d 明杠: %d 暗杠: %d\n", p, ret[p].MingGangNum, ret[p].AnGangNum)

	}
	return ret
}

func (ps MatchStats) Round() int {
	round := 0
	for _, r := range ps {
		if l := len(r); l > round {
			round = l
		}
	}

	return round
}
