package desk

import (
	"net/http"
	"strconv"

	"github.com/lonnng/nanoserver/internal/whitelist"

	"github.com/lonnng/nanoserver/internal/encoding"
	"github.com/lonnng/nanoserver/internal/errutil"
	"github.com/lonnng/nanoserver/internal/protocol"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

func MakeHandler(ctx context.Context, s Service) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encoding.EncodeError),
	}

	deskList := kithttp.NewServer(
		ctx,
		makeDeskListEndpoint(s),
		decodeDeskListRequest,
		encoding.EncodeResponse,
		opts...)

	deskByID := kithttp.NewServer(
		ctx,
		makeDeskByIDEndpoint(s),
		decodeDeskByIDRequest,
		encoding.EncodeResponse,
		opts...)

	deleteDeskByID := kithttp.NewServer(
		ctx,
		makeDeleteDeskEndpoint(s),
		decodeDeleteDeskRequest,
		encoding.EncodeResponse,
		opts...)

	getOptionsHandler := kithttp.NewServer(
		ctx,
		makeGetOptionsEndpoint(s),
		decodeGetOptionsRequest,
		encoding.EncodeResponse,
		opts...)

	r := mux.NewRouter()
	r.Handle("/v1/desk/{id}", deleteDeskByID).Methods("DELETE") //删除desk
	r.Handle("/v1/desk/player/{id}", deskList).Methods("GET")   //获取desk列表(lite)
	r.Handle("/v1/desk/{id}", deskByID).Methods("GET")          //获取desk记录
	r.Handle("/v1/desk/", getOptionsHandler).Methods("OPTIONS") //获取可用操作

	return r
}

func decodeDeleteDeskRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !whitelist.VerifyIP(r.RemoteAddr) {
		return nil, errutil.YXErrPermissionDenied
	}
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok || id == "" {
		return nil, errutil.YXErrInvalidParameter
	}
	return &protocol.DeleteProductionRequest{ProductionID: id}, nil
}

func decodeDeskListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	//if auth := r.Header.Get("Authorization"); auth == "" || !cache.Exists(auth) {
	//	return nil, errutil.YXErrAuthFailed
	//}

	if !whitelist.VerifyIP(r.RemoteAddr) {
		return nil, errutil.YXErrPermissionDenied
	}
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		return nil, errutil.YXErrInvalidParameter
	}

	id, err := strconv.ParseInt(idStr, 10, 0)
	if err != nil {
		return nil, errutil.YXErrInvalidParameter
	}

	return &protocol.DeskListRequest{
		Player: id,
	}, nil

}

func decodeDeskByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !whitelist.VerifyIP(r.RemoteAddr) {
		return nil, errutil.YXErrPermissionDenied
	}
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		return nil, errutil.YXErrInvalidParameter
	}

	id, err := strconv.ParseInt(idStr, 10, 0)
	if err != nil {
		return nil, errutil.YXErrInvalidParameter
	}

	return &protocol.DeskByIDRequest{ID: id}, nil
}

func decodeGetOptionsRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return &protocol.EmptyRequest{}, nil
}
