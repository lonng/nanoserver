package db

import (
	"time"
	log "github.com/sirupsen/logrus"

	"github.com/lonnng/nanoserver/internal/errutil"
	"github.com/lonnng/nanoserver/db/model"
)

func InsertOnline(count int, deskCount int) {
	o := model.Online{
		Time:      time.Now().Unix(),
		UserCount: count,
		DeskCount: deskCount,
	}

	_, err := DB.Insert(o)
	if err != nil {
		log.Errorf("统计在线人数失败: %s", err.Error())
	}
}

func OnlineStats(begin, end int64) ([]model.Online, error) {
	if begin > end {
		return nil, errutil.YXErrIllegalParameter
	}

	list := []model.Online{}

	return list, DB.Where("`time` BETWEEN ? AND ?", begin, end).Find(&list)
}
