package order

import (
	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"

	"github.com/lonnng/nanoserver/internal/errutil"
	"github.com/lonnng/nanoserver/internal/protocol"
	"github.com/lonnng/nanoserver/internal/types"
)

func makeCreateOrderEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r, ok := request.(*protocol.CreateOrderRequest)
		if !ok {
			return nil, errutil.YXErrWrongType
		}

		// todo:  check user exists
		/*if !cache.Exists(r.Token) {
			return nil, errutil.YXErrTokenNotFound
		}*/
		return s.CreateOrder(r)
	}
}

func makeYXPayOrderListEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		if r, ok := request.(*protocol.PayOrderListRequest); ok {
			list, total, err := s.YXPayOrderList(r)
			if err != nil {
				return nil, err
			}
			return protocol.PayOrderListResponse{
				Data:  list,
				Total: total,
			}, nil
		}
		return nil, errutil.YXErrWrongType
	}
}

func makeOrderListEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		if r, ok := request.(*protocol.OrderListRequest); ok {
			list, total, err := s.OrderList(r)
			if err != nil {
				return nil, err
			}
			return protocol.OrderListResponse{
				Data:  list,
				Total: total,
			}, nil
		}
		return nil, errutil.YXErrWrongType
	}
}

func makeTradeListEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		if r, ok := request.(*protocol.TradeListRequest); ok {
			list, total, err := s.TradeList(r)
			if err != nil {
				return nil, err
			}
			return protocol.TradeListResponse{
				Data:  list,
				Total: total,
			}, nil
		}
		return nil, errutil.YXErrWrongType
	}
}

func makeBalanceListEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		if r, ok := request.(*protocol.BalanceListRequest); ok {
			m, err := s.BalanceList(r.Uids)
			if err != nil {
				return nil, err
			}
			return protocol.BalanceListResponse{
				Data: m,
			}, nil
		}
		return nil, errutil.YXErrWrongType
	}
}

func makeWechatOrderCallbackEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r, ok := request.(*protocol.WechatOrderCallbackRequest)
		if !ok {
			return nil, errutil.YXErrWrongType
		}

		resp, err := s.Notify(types.YX, &protocol.UnifyOrderCallbackRequest{
			PayPlatform: types.Wechat,
			RawRequest:  r,
		})

		if err != nil {
			return protocol.ErrorResponse{Error: err.Error()}, nil
		}
		return resp, nil
	}
}

func makeGetOptionsEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return protocol.StringResponse{Data: s.GetOptions()}, nil
	}
}
