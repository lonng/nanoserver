package db

import (
	"time"

	"github.com/lonnng/nanoserver/db/model"
	"github.com/lonnng/nanoserver/internal/errutil"
)

func InsertDesk(h *model.Desk) error {
	if h == nil {
		return errutil.YXErrInvalidParameter
	}
	_, err := DB.Insert(h)
	if err != nil {
		return err
	}
	return nil
}

func UpdateDesk(d *model.Desk) error {
	_, err := DB.Exec("UPDATE `desk` SET `score_change0` = ?, `score_change1` = ?, `score_change2` = ?, `score_change3` = ?, `round` = ?  WHERE `id`= ? ",
		d.ScoreChange0,
		d.ScoreChange1,
		d.ScoreChange2,
		d.ScoreChange3,
		d.Round,
		d.Id)
	if err != nil {
		return err
	}

	//更新桌子时的同时,为排行准备数据
	f := func(uid int64, score int64, name string) error {
		if uid < 1 {
			return nil
		}
		r := &model.Rank{
			Uid:      uid,
			Score:    score,
			Match:    int64(d.Round),
			RecordAt: time.Now().Unix(),
			Name:     name,
			ClubId:   d.ClubId,
			DeskNo:   d.DeskNo,
			Creator:  d.Creator,
		}

		return InsertRank(r)
	}

	if d.Round > 0 {
		f(d.Player0, int64(d.ScoreChange0), d.PlayerName0)
		f(d.Player1, int64(d.ScoreChange1), d.PlayerName1)
		f(d.Player2, int64(d.ScoreChange2), d.PlayerName2)
		f(d.Player3, int64(d.ScoreChange3), d.PlayerName3)
	}

	return nil
}

func QueryDesk(id int64) (*model.Desk, error) {
	h := &model.Desk{Id: id}
	has, err := DB.Get(h)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errutil.YXErrDeskNotFound
	}
	return h, nil
}

//指定的桌子是否存在
func DeskNumberExists(no string) bool {
	d := &model.Desk{
		DeskNo: no,
	}

	has, err := DB.Get(d)
	if err != nil {
		return true
	}
	return has
}

func DeleteDesk(id int64) error {
	_, err := DB.Delete(&model.Desk{Id: id})
	return err
}

func DeskList(player int64, offset, count int) ([]model.Desk, int, error) {
	const (
		limit = 15
	)
	result := make([]model.Desk, 0)
	err := DB.Where("(player0 = ? OR player1 = ? OR player2 = ? OR player3 = ? ) AND round > 0",
		player, player, player, player).Desc("created_at").Limit(limit, 0).Find(&result)

	if err != nil {
		return nil, 0, errutil.YXErrDBOperation
	}
	return result, len(result), nil
}
