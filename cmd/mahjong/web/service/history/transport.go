package history

import (
	"net/http"
	"strconv"

	"github.com/lonnng/nanoserver/internal/whitelist"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"github.com/lonnng/nanoserver/internal/cache"
	"github.com/lonnng/nanoserver/internal/encoding"
	"github.com/lonnng/nanoserver/internal/errutil"
	"github.com/lonnng/nanoserver/internal/protocol"
)

func MakeHandler(ctx context.Context, s Service) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encoding.EncodeError),
	}

	historyLiteList := kithttp.NewServer(
		ctx,
		makeHistoryLiteListEndpoint(s),
		decodeHistoryLiteListRequest,
		encoding.EncodeResponse,
		opts...)

	historyByID := kithttp.NewServer(
		ctx,
		makeHistoryByIDEndpoint(s),
		decodeHistoryByIDRequest,
		encoding.EncodeResponse,
		opts...)

	deleteHistoryByID := kithttp.NewServer(
		ctx,
		makeDeleteHistoryEndpoint(s),
		decodeDeleteHistoryRequest,
		encoding.EncodeResponse,
		opts...)

	getOptionsHandler := kithttp.NewServer(
		ctx,
		makeGetOptionsEndpoint(s),
		decodeGetOptionsRequest,
		encoding.EncodeResponse,
		opts...)

	r := mux.NewRouter()
	r.Handle("/v1/history/{id}", deleteHistoryByID).Methods("DELETE")      //删除产品
	r.Handle("/v1/history/lite/{desk_id}", historyLiteList).Methods("GET") //获取历史列表(lite),参数为deskid
	r.Handle("/v1/history/{id}", historyByID).Methods("GET")               //获取历史记录
	r.Handle("/v1/history/", getOptionsHandler).Methods("OPTIONS")         //获取可用操作

	return r
}

func decodeDeleteHistoryRequest(_ context.Context, r *http.Request) (interface{}, error) {
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

func decodeHistoryLiteListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	//if auth := r.Header.Get("Authorization"); auth == "" || !cache.Exists(auth) {
	//	return nil, errutil.YXErrAuthFailed
	//}

	if !whitelist.VerifyIP(r.RemoteAddr) {
		return nil, errutil.YXErrPermissionDenied
	}
	vars := mux.Vars(r)
	idStr, ok := vars["desk_id"]
	if !ok || idStr == "" {
		return nil, errutil.YXErrInvalidParameter
	}

	id, err := strconv.ParseInt(idStr, 10, 0)
	if err != nil {
		return nil, errutil.YXErrInvalidParameter
	}

	return &protocol.HistoryLiteListRequest{DeskID: id}, nil

}

func decodeHistoryListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if auth := r.Header.Get("Authorization"); auth == "" || !cache.Exists(auth) {
		return nil, errutil.YXErrAuthFailed
	}
	return &protocol.EmptyRequest{}, nil
}

func decodeHistoryByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
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

	return &protocol.HistoryByIDRequest{ID: id}, nil
}

func decodeGetOptionsRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return &protocol.EmptyRequest{}, nil
}
