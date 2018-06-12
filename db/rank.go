package db

import (
	log "github.com/sirupsen/logrus"
	"github.com/lonnng/nanoserver/internal/errutil"
	"github.com/lonnng/nanoserver/internal/protocol"
	"time"

	"fmt"
	"github.com/lonnng/nanoserver/db/model"
	"strconv"
)

func InsertRank(r *model.Rank) error {
	if r == nil {
		return errutil.YXErrInvalidParameter
	}
	_, err := DB.Insert(r)
	if err != nil {
		return err
	}

	return nil
}

func rankHelper(typ, n int, from int64) ([]protocol.Rank, error) {
	order := "DESC"
	compr := ">"
	if typ == RankingDesc {
		order = ""
		compr = "<"
	}

	sql := fmt.Sprintf("SELECT * FROM (SELECT `name`, `uid`, SUM(`score`) AS `value` FROM rank"+
		" WHERE `record_at` BETWEEN %d AND %d GROUP BY `uid`, `name` ORDER BY `value` %s limit %d) as ranking WHERE `value` %s 0;",
		from,
		time.Now().Unix(),
		order,
		n,
		compr)

	mQuery, err := DB.Query(sql)

	if len(mQuery) < 1 || err != nil {
		log.Error(err)
		return nil, err
	}

	var ret []protocol.Rank

	for _, m := range mQuery {
		r := protocol.Rank{
			Name: string(m["name"]),
		}

		temp := string(m["uid"])
		if temp != "" {
			r.Uid, err = strconv.ParseInt(temp, 10, 0)
		}

		temp = string(m["value"])
		if temp != "" {
			r.Value, err = strconv.ParseInt(temp, 10, 0)
		}

		ret = append(ret, r)
	}

	return ret, nil
}

//分数、对局排行
func RankList(typ, n int, date int64) ([][]protocol.Rank, error) {
	now := time.Now()

	//日/周/月
	t1 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	t2 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	t2 = time.Date(now.Year(), now.Month(), now.Day()-int(t2.Weekday()), 0, 0, 0, 0, time.Local)
	t3 := time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, time.Local)

	ts := []int64{t1.Unix(), t2.Unix(), t3.Unix()}

	var list [][]protocol.Rank

	for _, t := range ts {
		r, err := rankHelper(typ, n, t)
		if err != nil {
			log.Error(err)
			r = []protocol.Rank{}
		}

		list = append(list, r)
	}

	return list, nil
}
