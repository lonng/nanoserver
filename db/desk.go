package db

import (
	"github.com/lonng/nanoserver/db/model"
	"github.com/lonng/nanoserver/pkg/errutil"
)

func InsertDesk(h *model.Desk) error {
	if h == nil {
		return errutil.ErrInvalidParameter
	}
	_, err := database.Insert(h)
	if err != nil {
		return err
	}
	return nil
}

func UpdateDesk(d *model.Desk) error {
	_, err := database.Exec("UPDATE `desk` SET `score_change0` = ?, `score_change1` = ?, `score_change2` = ?, `score_change3` = ?, `round` = ?  WHERE `id`= ? ",
		d.ScoreChange0,
		d.ScoreChange1,
		d.ScoreChange2,
		d.ScoreChange3,
		d.Round,
		d.Id)
	if err != nil {
		return err
	}
	return nil
}

func QueryDesk(id int64) (*model.Desk, error) {
	h := &model.Desk{Id: id}
	has, err := database.Get(h)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errutil.ErrDeskNotFound
	}
	return h, nil
}

//指定的桌子是否存在
func DeskNumberExists(no string) bool {
	d := &model.Desk{
		DeskNo: no,
	}

	has, err := database.Get(d)
	if err != nil {
		return true
	}
	return has
}

func DeleteDesk(id int64) error {
	_, err := database.Delete(&model.Desk{Id: id})
	return err
}

func DeskList(player int64) ([]model.Desk, int, error) {
	const (
		limit = 15
	)
	result := make([]model.Desk, 0)
	err := database.Where("(player0 = ? OR player1 = ? OR player2 = ? OR player3 = ? ) AND round > 0",
		player, player, player, player).Desc("created_at").Limit(limit, 0).Find(&result)

	if err != nil {
		return nil, 0, errutil.ErrDBOperation
	}
	return result, len(result), nil
}
