package db

import (
	"github.com/lonng/nanoserver/db/model"
	"github.com/lonng/nanoserver/protocol"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

func InsertConsume(entity *model.CardConsume) error {
	_, err := database.Insert(entity)
	if err != nil {
		log.Error(err)
	}

	return err
}

//消耗统计
func ConsumeStats(from, to int64) ([]*protocol.CardConsume, error) {
	fn := func(from, to int64) *protocol.CardConsume {
		mQuery, err := database.Query("SELECT SUM(card_count) AS cards FROM card_consume WHERE consume_at BETWEEN ? AND ?; ",
			from,
			to)

		cc := &protocol.CardConsume{
			Date: from,
		}

		if len(mQuery) < 1 || err != nil {
			return cc
		}

		temp := string(mQuery[0]["cards"])
		if temp != "" {
			cc.Value, err = strconv.ParseInt(temp, 10, 0)
		}

		return cc

	}
	begin := time.Unix(from, 0)

	var ret []*protocol.CardConsume

	t := time.Date(begin.Year(), begin.Month(), begin.Day(), 0, 0, 0, 0, time.Local)

	for i := t.Unix(); i < to; i += dayInSecond {
		cc := fn(i, i+dayInSecond-1)

		ret = append(ret, cc)
	}
	return ret, nil
}
