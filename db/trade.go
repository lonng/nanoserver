package db

import (
	"github.com/lonng/nanoserver/db/model"
	"github.com/lonng/nanoserver/pkg/algoutil"
	"github.com/lonng/nanoserver/pkg/errutil"
)

func InsertTrade(t *model.Trade) error {
	logger.Info("insert trade, order id: " + t.OrderId)

	trade := &model.Trade{OrderId: t.OrderId}
	has, err := database.Get(trade)
	if err != nil {
		return err
	}
	if has {
		return errutil.ErrTradeExisted
	}
	order, err := QueryOrder(t.OrderId)
	if err != nil {
		return err
	}
	if order.Type == OrderTypeBuyToken {
		order.Status = OrderStatusNotified
	} else {
		order.Status = OrderStatusPayed
	}
	sess := database.NewSession()

	// 开始事务
	sess.Begin()
	defer sess.Close()
	if _, err := sess.Insert(t); err != nil {
		println(err.Error())
		sess.Rollback()
		return err
	}

	if _, err := sess.Where("order_id = ?", order.OrderId).Update(order); err != nil {
		println(err.Error())
		sess.Rollback()
		return err
	}

	u := &model.User{}
	sess.Where("uid = ?", order.Uid).Get(u)

	//添加首充时间
	if u.FirstRechargeAt == 0 {
		u.FirstRechargeAt = order.CreatedAt
		if _, err = sess.Id(u.Id).Update(u); err != nil {
			sess.Rollback()
			return err
		}
	}

	return sess.Commit()
}

func TradeList(appid, channelID, orderID string, start, end int64, offset, count int) ([]ViewTrade, int, error) {
	start, end = algoutil.TimeRange(start, end)

	trade := &ViewTrade{
		AppId:     appid,
		ChannelId: channelID,
		OrderId:   orderID,
	}
	total, err := database.Where("pay_at BETWEEN ? AND ?", start, end).Count(trade)
	if err != nil {
		logger.Error(err)
		return nil, 0, errutil.ErrDBOperation
	}

	result := make([]ViewTrade, 0)
	if count == noLimitFlag {
		err = database.Where("pay_at BETWEEN ? AND ?", start, end).
			Desc("id").Find(&result, trade)
	} else {
		err = database.Where("pay_at BETWEEN ? AND ?", start, end).
			Desc("id").Limit(count, offset).Find(&result, trade)
	}

	if err != nil {
		logger.Error(err)
		return nil, 0, errutil.ErrDBOperation
	}

	return result, int(total), nil
}
