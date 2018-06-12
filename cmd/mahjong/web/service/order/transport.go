package order

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"

	"github.com/lonnng/nanoserver/internal/algoutil"
	"github.com/lonnng/nanoserver/internal/encoding"
	"github.com/lonnng/nanoserver/internal/errutil"
	"github.com/lonnng/nanoserver/internal/protocol"
	"github.com/lonnng/nanoserver/internal/whitelist"
)

func MakeHandler(ctx context.Context, s Service) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encoding.EncodeError),
	}
	createOrderHandler := kithttp.NewServer(
		ctx,
		makeCreateOrderEndpoint(s),
		decodeCreateOrderRequest,
		encoding.EncodeResponse,
		opts...)

	yxPayOrderStatsListHandler := kithttp.NewServer(
		ctx,
		makeYXPayOrderListEndpoint(s),
		decodeYXPayOrderListEndpoint,
		encoding.EncodeResponse,
		opts...)

	orderListHandler := kithttp.NewServer(
		ctx,
		makeOrderListEndpoint(s),
		decodeOrderListEndpoint,
		encoding.EncodeResponse,
		opts...)

	tradeListHandler := kithttp.NewServer(
		ctx,
		makeTradeListEndpoint(s),
		decodeTradeListEndpoint,
		encoding.EncodeResponse,
		opts...)

	wechatOrderCallbackHandler := kithttp.NewServer(
		ctx,
		makeWechatOrderCallbackEndpoint(s),
		decodeWechatOrderCallbackRequest,
		encoding.EncodeResponse,
		opts...)

	getOptionsHandler := kithttp.NewServer(
		ctx,
		makeGetOptionsEndpoint(s),
		decodeGetOptionsRequest,
		encoding.EncodeResponse,
		opts...)

	r := mux.NewRouter()

	r.Handle("/v1/order/console/stats/snakepay/", yxPayOrderStatsListHandler).Methods("GET") //蛇币的收支统计
	r.Handle("/v1/order/console/", orderListHandler).Methods("GET")                          //订单列表
	r.Handle("/v1/order/console/trade/", tradeListHandler).Methods("GET")                    //交易列表

	r.Handle("/v1/order/", createOrderHandler).Methods("GET")                       //创建订单
	r.Handle("/v1/order/notify/wechat", wechatOrderCallbackHandler).Methods("POST") //微信订单回调

	r.Handle("/v1/order/", getOptionsHandler).Methods("OPTIONS") //获取可用操作
	return r
}

func decodeCreateOrderRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	count, err := strconv.Atoi(strings.TrimSpace(r.Form.Get("count")))
	if err != nil {
		return nil, err
	}
	uid, err := strconv.ParseInt(strings.TrimSpace(r.Form.Get("uid")), 10, 64)
	if err != nil {
		return nil, err
	}
	appId := strings.TrimSpace(r.Form.Get("appId"))
	channelId := strings.TrimSpace(r.Form.Get("channelId"))
	platform := strings.TrimSpace(r.Form.Get("platform"))
	extra := strings.TrimSpace(r.Form.Get("extra"))
	name := strings.TrimSpace(r.Form.Get("name"))
	if appId == "" || channelId == "" || platform == "" || count == 0 || name == "" {
		return nil, errutil.YXErrIllegalParameter
	}

	return &protocol.CreateOrderRequest{
		AppID:          appId,
		ChannelID:      channelId,
		Platform:       platform,
		ProductionName: name,
		ProductCount:   count,
		Extra:          extra,
		Uid:            uid,
		Device:         protocol.Device{Remote: r.RemoteAddr},
	}, nil
}

//订单列表
func decodeOrderListEndpoint(_ context.Context, r *http.Request) (interface{}, error) {
	if !whitelist.VerifyIP(r.RemoteAddr) {
		return nil, errutil.YXErrPermissionDenied
	}

	request := &protocol.OrderListRequest{}
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	request.Offset = algoutil.RetriveIntOrDefault(r.Form.Get("offset"), 0)
	request.Count = algoutil.RetriveIntOrDefault(r.Form.Get("count"), -1)
	request.PayBy = strings.ToLower(r.Form.Get("pay_by"))
	request.Status = uint8(algoutil.RetriveIntOrDefault(r.Form.Get("status"), 0))
	request.AppID = r.Form.Get("appid")
	request.ChannelID = r.Form.Get("channel_id")
	request.Start = algoutil.RetriveInt64OrDefault(r.Form.Get("start"), -1)
	request.End = algoutil.RetriveInt64OrDefault(r.Form.Get("end"), -1)
	request.Uid = r.Form.Get("uid")
	request.OrderID = r.Form.Get("order_id")

	return request, nil
}

//交易列表
func decodeTradeListEndpoint(_ context.Context, r *http.Request) (interface{}, error) {
	if !whitelist.VerifyIP(r.RemoteAddr) {
		return nil, errutil.YXErrPermissionDenied
	}

	request := &protocol.TradeListRequest{}
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	request.Offset = algoutil.RetriveIntOrDefault(r.Form.Get("offset"), 0)
	request.Count = algoutil.RetriveIntOrDefault(r.Form.Get("count"), -1)
	request.AppID = r.Form.Get("appid")
	request.ChannelID = r.Form.Get("channel_id")
	request.Start = algoutil.RetriveInt64OrDefault(r.Form.Get("start"), -1)
	request.End = algoutil.RetriveInt64OrDefault(r.Form.Get("end"), -1)
	request.OrderID = r.Form.Get("order_id")

	return request, nil
}

//悠闲币的收支列表
func decodeYXPayOrderListEndpoint(_ context.Context, r *http.Request) (interface{}, error) {
	if !whitelist.VerifyIP(r.RemoteAddr) {
		return nil, errutil.YXErrPermissionDenied
	}

	request := &protocol.PayOrderListRequest{}
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	request.Offset = algoutil.RetriveIntOrDefault(r.Form.Get("offset"), 0)
	request.Count = algoutil.RetriveIntOrDefault(r.Form.Get("count"), -1)
	request.Type = algoutil.RetriveIntOrDefault(r.Form.Get("type"), 0)

	request.AppID = r.Form.Get("appid")
	request.ChannelID = r.Form.Get("channel_id")
	request.Start = algoutil.RetriveInt64OrDefault(r.Form.Get("start"), -1)
	request.End = algoutil.RetriveInt64OrDefault(r.Form.Get("end"), -1)
	request.Uid = algoutil.RetriveInt64OrDefault(r.Form.Get("uid"), 0)

	return request, nil
}

func decodeBalanceListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !whitelist.VerifyIP(r.RemoteAddr) {
		return nil, errutil.YXErrPermissionDenied
	}
	request := &protocol.BalanceListRequest{}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeObtainBalanceRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var auth string
	if auth = r.Header.Get("Authorization"); auth == "" {
		return nil, errutil.YXErrAuthFailed
	}
	return &protocol.ObtainBalanceReqeust{Token: auth}, nil
}

func decodeWechatOrderCallbackRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &protocol.WechatOrderCallbackRequest{}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if err := xml.Unmarshal(data, request); err != nil {
		return nil, err
	}

	request.Raw = string(data)

	return request, nil
}

func decodeGetOptionsRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return &protocol.EmptyRequest{}, nil
}
