package yx

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
	"github.com/lonnng/nanoserver/cmd/mahjong/web/provider/yxsdk"
	"github.com/lonnng/nanoserver/db"
	"github.com/lonnng/nanoserver/db/model"
	"github.com/lonnng/nanoserver/internal/errutil"
	"github.com/lonnng/nanoserver/internal/protocol"
	"github.com/lonnng/nanoserver/internal/types"
)

var exchangeRate = 10

type YXPay struct{}

func (sp *YXPay) CreateOrderResponse(order *model.Order) (interface{}, error) {
	log.Infof("order_info %+v\n", order)
	if err := db.UserLoseCoinByUID(order.Uid, int64(order.Money)); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return protocol.CreateOrderSnakeResponse{
		Result:      "success",
		PayPlatform: types.YX,
	}, nil
}

func (sp *YXPay) Notify(request interface{}) (*model.Trade, interface{}, error) {
	if order, ok := request.(*model.Order); ok {
		t := time.Now().Unix()
		trade := &model.Trade{
			OrderId:     order.OrderId,
			PayOrderId:  order.OrderId,
			PayPlatform: types.YX,
			PayCreateAt: t,
			PayAt:       t,
			ComsumerId:  fmt.Sprintf("%d", order.Uid),
			MerchantId:  types.YX,
		}
		return trade, protocol.SuccessResponse, nil
	}
	return nil, nil, errutil.YXErrWrongType
}

func (sp *YXPay) Setup() error {
	log.Info("pay_provider: yx")

	e := viper.GetInt("exchange_rate")
	if e < 0 {
		return errors.New("exchange rate can not less than zero")
	}
	exchangeRate = e
	return nil
}

func init() {
	yxsdk.Register(types.YX, &YXPay{})
}
