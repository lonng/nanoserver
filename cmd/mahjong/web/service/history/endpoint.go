package history

import (
	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"

	"github.com/lonnng/nanoserver/internal/errutil"
	"github.com/lonnng/nanoserver/internal/protocol"
)


func makeHistoryLiteListEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r, ok := request.(*protocol.HistoryLiteListRequest)
		if !ok {
			return nil, errutil.YXErrWrongType
		}

		list, t, err := s.HistoryLiteList(r)
		if err != nil {
			return nil, err
		}
		return protocol.HistoryLiteListResponse{Data: list, Total: t}, nil
	}
}

func makeHistoryListEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r, ok := request.(*protocol.HistoryListRequest)
		if !ok {
			return nil, errutil.YXErrWrongType
		}

		list, t, err := s.HistoryList(r)
		if err != nil {
			return nil, err
		}
		return protocol.HistoryListResponse{Data: list, Total: t}, nil
	}
}

func makeHistoryByIDEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r, ok := request.(*protocol.HistoryByIDRequest)
		if !ok {
			return nil, errutil.YXErrWrongType
		}

		h, err := s.HistoryByID(r.ID)
		if err != nil {
			return nil, err
		}
		return protocol.HistoryByIDResponse{
			Data: h,
		}, nil
	}
}


func makeDeleteHistoryEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		if r, ok := request.(*protocol.DeleteHistoryRequest); ok {
			if err := s.DeleteHistory(r.ID); err != nil {
				return nil, err
			}
			return protocol.SuccessResponse, nil
		}
		return nil, errutil.YXErrWrongType
	}
}

func makeGetOptionsEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return protocol.StringResponse{Data: s.GetOptions()}, nil
	}
}
