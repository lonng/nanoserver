package desk

import (
	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"

	"github.com/lonnng/nanoserver/internal/errutil"
	"github.com/lonnng/nanoserver/internal/protocol"

)


func makeDeskListEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r, ok := request.(*protocol.DeskListRequest)
		if !ok {
			return nil, errutil.YXErrWrongType
		}

		list, t, err := s.DeskList(r)
		if err != nil {
			return nil, err
		}
		return protocol.DeskListResponse{Data: list, Total: t}, nil
	}
}


func makeDeskByIDEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r, ok := request.(*protocol.DeskByIDRequest)
		if !ok {
			return nil, errutil.YXErrWrongType
		}

		h, err := s.DeskByID(r.ID)
		if err != nil {
			return nil, err
		}
		return protocol.DeskByIDResponse{
			Data: h,
		}, nil
	}
}


func makeDeleteDeskEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		if r, ok := request.(*protocol.DeleteDeskByIDRequest); ok {
			if err := s.DeleteDesk(r.ID); err != nil {
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
