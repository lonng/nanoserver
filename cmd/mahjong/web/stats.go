package web

import (
	"github.com/lonng/nex"
	log "github.com/sirupsen/logrus"

	"time"

	"github.com/lonng/nanoserver/db"
	"github.com/lonng/nanoserver/protocol"

	"github.com/lonng/nanoserver/pkg/errutil"
)

var dayInternal = 24 * 60 * 60

//注册用户数
func registerUsersHandler(query *nex.Form) (interface{}, error) {
	begin := query.Int64OrDefault("from", 0)
	end := query.Int64OrDefault("to", -1)
	if end < 0 {
		end = time.Now().Unix()
	}

	log.Infof("获取注册用户信息: begin=%s, end=%s", time.Unix(begin, 0).String(), time.Unix(end, 0).String())
	c, err := db.QueryRegisterUsers(begin, end)
	if err != nil {
		return nil, err
	}

	return protocol.CommonResponse{
		Data: c,
	}, nil
}

//实时在线人数
func onlineLiteHandler() (interface{}, error) {

	c, err := db.OnlineStatsLite()
	if err != nil {
		return nil, err
	}

	return protocol.CommonResponse{
		Data: c,
	}, nil
}

//某日的n日留存
func retentionHandler(query *nex.Form) (interface{}, error) {
	from := query.IntOrDefault("from", -1)
	to := query.IntOrDefault("to", -1)

	if from < 0 || to < 0 || to < from {
		return nil, errutil.ErrIllegalParameter
	}

	list := []*protocol.Retention{}
	for i := from; i <= to; i += dayInternal {
		ret, err := db.RetentionList(i)
		if err != nil {
			return nil, err
		}
		list = append(list, ret)
	}

	return &protocol.RetentionResponse{Data: list}, nil

}

//从指定日开始到当前的每日房卡消耗
func cardConsumeStatsHandler(query *nex.Form) (interface{}, error) {
	from := query.Int64OrDefault("from", 0)
	to := query.Int64OrDefault("to", 0)

	if to == 0 {
		from = time.Now().Unix()
	}

	ret, err := db.ConsumeStats(from, to)
	if err != nil {
		return nil, err
	}

	return &protocol.RetentionResponse{Data: ret}, nil

}

//从指定日开始到当前的每日活跃人数
func activationUsersHandler(query *nex.Form) (interface{}, error) {
	from := query.Int64OrDefault("from", 0)
	to := query.Int64OrDefault("to", 0)

	if to == 0 {
		to = time.Now().Unix()
	}

	ret, err := db.QueryActivationUser(from, to)
	if err != nil {
		return nil, err
	}

	return &protocol.RetentionResponse{Data: ret}, nil

}
