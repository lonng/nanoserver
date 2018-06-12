package db

import (
	"github.com/lonnng/nanoserver/internal/errutil"
	log "github.com/sirupsen/logrus"

	"github.com/lonnng/nanoserver/db/model"
)

func InsertHistory(h *model.History) error {
	if h == nil {
		return errutil.YXErrInvalidParameter
	}
	_, err := DB.Insert(h)
	if err != nil {
		return errutil.YXErrDBOperation
	}
	return nil
}

func QueryHistory(id int64) (*model.History, error) {
	h := &model.History{Id: id}
	has, err := DB.Get(h)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if !has {
		return nil, errutil.YXErrOrderNotFound
	}
	return h, nil
}

func DeleteHistory(id int64) error {
	_, err := DB.Delete(&model.History{Id: id})
	return err
}

func DeleteHistoriesByDeskID(deskId int64) error {
	_, err := DB.Delete(&model.History{DeskId: deskId})
	return err
}

func QueryHistoriesByDeskID(deskID int64) ([]model.History, int, error) {
	result := make([]model.History, 0)
	err := DB.Where("desk_id=?", deskID).Asc("begin_at").Find(&result)
	if err != nil {
		log.Error(err)
		return nil, 0, errutil.YXErrDBOperation
	}
	return result, len(result), nil
}


