package test

import (
	"fmt"
	"time"

	"github.com/spf13/viper"

	"github.com/lonnng/nanoserver/internal/errutil"
	"github.com/lonnng/nanoserver/internal/protocol"
	"github.com/lonnng/nanoserver/internal/types"

	"github.com/lonnng/nanoserver/cmd/mahjong/web/provider/yxsdk"
	"github.com/lonnng/nanoserver/db/model"

	log "github.com/sirupsen/logrus"
)

type TestPay struct {
	enablePayTest bool
}

func (tp *TestPay) CreateOrderResponse(order *model.Order) (interface{}, error) {
	if !tp.enablePayTest {
		return nil, errutil.YXErrPayTestDisable
	}

	log.Infof("order_info: %+v", order)
	return protocol.CreateOrderSnakeResponse{
		Result:      "success",
		PayPlatform: types.Test,
	}, nil
}

func (tp *TestPay) Notify(request interface{}) (*model.Trade, interface{}, error) {
	order, ok := request.(*model.Order)

	if !ok {
		return nil, nil, errutil.YXErrWrongType
	}

	t := time.Now().Unix()
	trade := &model.Trade{
		OrderId:     order.OrderId,
		PayOrderId:  order.OrderId,
		PayPlatform: types.Test,
		PayCreateAt: t,
		PayAt:       t,
		ComsumerId:  fmt.Sprintf("%d", order.Uid),
		MerchantId:  types.Test,
	}
	return trade, protocol.SuccessResponse, nil
}

func (tp *TestPay) Setup() error {
	log.Info("pay_provider: test")
	tp.enablePayTest = viper.GetBool("enable_pay_test")

	log.Infof("enable_pay_test: %t", tp.enablePayTest)
	if !tp.enablePayTest {
		yxsdk.Remove(types.Test)
	}

	return nil
}

func init() {
	yxsdk.Register(types.Test, &TestPay{})
}
