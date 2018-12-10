package db

import (
	"strings"

	"github.com/lonng/nanoserver/db/model"
	"github.com/lonng/nanoserver/pkg/algoutil"
	"github.com/lonng/nanoserver/pkg/errutil"
)

const (
	noLimitFlag  = -1 //如果count == -1则表示返回所有数据
	noTimeFilter = -1 //如果start/end == -1则表示无时间筛选
)

func QueryOrder(orderID string) (*model.Order, error) {
	order := &model.Order{OrderId: orderID}
	has, err := database.Get(order)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errutil.ErrOrderNotFound
	}
	return order, nil
}

func InsertOrder(order *model.Order) error {
	if order == nil {
		return errutil.ErrInvalidParameter
	}
	_, err := database.Insert(order)
	if err != nil {
		return errutil.ErrDBOperation
	}
	return nil
}

func YXPayOrderList(uid int64, appid, channelID, orderID string, start, end int64, typ, offset, count int) ([]model.Order, int, error) {

	order := &model.Order{
		AppId:     appid,
		ChannelId: channelID,
		Uid:       uid,
		OrderId:   orderID,
		Type:      typ,
	}

	start, end = algoutil.TimeRange(start, end)

	//println("uid", uid, "appid", appid, "channelid", channelID, "start", start, "end", end, "offset", offset, "count", count)

	total, err := database.Where("created_at BETWEEN ? AND ?", start, end).Count(order)
	if err != nil {
		logger.Error(err)
		return nil, 0, errutil.ErrDBOperation
	}

	result := make([]model.Order, 0)
	if count == noLimitFlag {
		err = database.Where("created_at BETWEEN ? AND ?", start, end).
			Desc("id").Find(&result, order)
	} else {
		err = database.Where("created_at BETWEEN ? AND ?", start, end).
			Desc("id").Limit(count, offset).Find(&result, order)
	}

	if err != nil {
		logger.Error(err)
		return nil, 0, errutil.ErrDBOperation
	}

	return result, int(total), nil
}

func OrderList(uid int64, appid, channelID, orderID, payBy string, start, end int64, status, offset, count int) ([]model.Order, int, error) {
	order := &model.Order{
		AppId:       appid,
		ChannelId:   channelID,
		Uid:         uid,
		OrderId:     orderID,
		PayPlatform: payBy,
		Status:      status,
	}

	start, end = algoutil.TimeRange(start, end)

	//println("uid", uid, "appid", appid, "channelid", channelID, "start", start, "end", end, "offset", offset, "count", count)

	total, err := database.Where("created_at BETWEEN ? AND ?", start, end).Count(order)
	if err != nil {
		logger.Error(err)
		return nil, 0, errutil.ErrDBOperation
	}

	result := make([]model.Order, 0)
	if count == noLimitFlag {
		err = database.Where("created_at BETWEEN ? AND ?", start, end).
			Desc("id").Find(&result, order)
	} else {
		err = database.Where("created_at BETWEEN ? AND ?", start, end).
			Desc("id").Limit(count, offset).Find(&result, order)
	}

	if err != nil {
		logger.Error(err)
		return nil, 0, errutil.ErrDBOperation
	}

	return result, int(total), nil
}

func BalanceList(uids []string) (map[string]string, error) {
	if uids == nil {
		return nil, errutil.ErrIllegalParameter
	}

	sql := "SELECT  uid, coin from `user` WHERE uid IN ( " + strings.Join(uids, ",") + ")"
	results, err := database.Query(sql)
	if err != nil {
		logger.Error(err)
		return nil, errutil.ErrDBOperation
	}

	m := make(map[string]string)

	for _, result := range results {
		m[string(result["uid"])] = string(result["coin"])
	}
	return m, nil

}
