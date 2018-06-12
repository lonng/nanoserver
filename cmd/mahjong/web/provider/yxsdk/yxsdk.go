package yxsdk

import (
	"github.com/go-kit/kit/log"

	"github.com/lonnng/nanoserver/internal/errutil"
	"github.com/lonnng/nanoserver/internal/protocol"
	"github.com/lonnng/nanoserver/internal/types"

	"github.com/lonnng/nanoserver/cmd/mahjong/web/provider"
	"github.com/lonnng/nanoserver/db/model"
)

type yxSDKProvider struct {
	provider.Base
	logger log.Logger
}

func (s *yxSDKProvider) CreateOrderResponse(order *model.Order) (interface{}, error) {
	p := pm.provider(order.PayPlatform)
	if p == nil {
		s.logger.Log("err", errutil.YXErrInvalidPayPlatform.Error())
		return nil, errutil.YXErrInvalidPayPlatform
	}

	return p.CreateOrderResponse(order)
}

func (s *yxSDKProvider) Notify(request interface{}) (*model.Trade, interface{}, error) {
	r, ok := request.(*protocol.UnifyOrderCallbackRequest)
	if !ok {
		return nil, nil, errutil.YXErrWrongType
	}

	p := pm.provider(r.PayPlatform)
	if p == nil {
		s.logger.Log("err", errutil.YXErrInvalidPayPlatform.Error())
		return nil, nil, errutil.YXErrInvalidPayPlatform
	}

	return p.Notify(r.RawRequest)
}

func (s *yxSDKProvider) Setup(logger log.Logger) error {
	s.logger = log.NewContext(logger).With("provider", "yxsdk")
	if err := pm.setup(logger); err != nil {
		panic(err)
	}
	s.logger.Log("msg", "running")
	return nil
}

func init() {
	provider.Register(types.YX, &yxSDKProvider{})
}
