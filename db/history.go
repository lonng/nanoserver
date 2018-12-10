package db

import (
	"github.com/lonng/nanoserver/pkg/errutil"
	log "github.com/sirupsen/logrus"

	"github.com/lonng/nanoserver/db/model"
)

func InsertHistory(h *model.History) error {
	if h == nil {
		return errutil.ErrInvalidParameter
	}
	_, err := database.Insert(h)
	if err != nil {
		return errutil.ErrDBOperation
	}
	return nil
}

func QueryHistory(id int64) (*model.History, error) {
	h := &model.History{Id: id}
	has, err := database.Get(h)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if !has {
		return nil, errutil.ErrOrderNotFound
	}
	return h, nil
}

func DeleteHistory(id int64) error {
	_, err := database.Delete(&model.History{Id: id})
	return err
}

func DeleteHistoriesByDeskID(deskId int64) error {
	_, err := database.Delete(&model.History{DeskId: deskId})
	return err
}

func QueryHistoriesByDeskID(deskID int64) ([]model.History, int, error) {
	result := make([]model.History, 0)
	err := database.Where("desk_id=?", deskID).Asc("begin_at").Find(&result)
	if err != nil {
		log.Error(err)
		return nil, 0, errutil.ErrDBOperation
	}
	return result, len(result), nil
}
