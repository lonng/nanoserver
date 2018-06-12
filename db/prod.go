package db

import (
	"github.com/lonnng/nanoserver/db/model"
	"github.com/lonnng/nanoserver/internal/errutil"
)

func QueryProduction(prodID string) (*model.Production, error) {
	p := &model.Production{ProductionId: prodID}
	has, err := DB.Get(p)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errutil.YXErrProductionNotFound
	}
	return p, nil
}

func InsertProduction(p *model.Production) error {
	_, err := DB.Insert(p)
	return err
}

func UpdateProduction(p *model.Production) error {
	_, err := DB.Where("production_id=?", p.ProductionId).Update(p)
	return err
}

//所有offline_at <>0 表示已经下线
func ProductionList(offset, count int) ([]model.Production, int64, error) {
	total, err := DB.Where("offline_at = 0").Count(&model.Production{})
	if err != nil {
		logger.Error(err)
		return nil, 0, errutil.YXErrDBOperation
	}

	// retrieve all record
	if count == -1 {
		count = int(total)
	}

	result := make([]model.Production, 0)
	err = DB.Where("offline_at = 0").Limit(count, offset).Find(&result)
	if err != nil {
		logger.Error(err)
		return nil, 0, errutil.YXErrDBOperation
	}
	return result, total, nil
}
