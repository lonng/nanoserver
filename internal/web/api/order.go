package api

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/lonng/nanoserver/db"
	"github.com/lonng/nanoserver/db/model"
	provider2 "github.com/lonng/nanoserver/internal/web/api/provider"
	"github.com/lonng/nex"
	"github.com/pborman/uuid"
	"golang.org/x/net/context"

	"github.com/lonng/nanoserver/pkg/errutil"
	"github.com/lonng/nanoserver/pkg/whitelist"
	"github.com/lonng/nanoserver/protocol"
)

func MakeOrderService() http.Handler {
	router := mux.NewRouter()
	router.Handle("/v1/order/console/", nex.Handler(orderList)).Methods("GET")            //订单列表
	router.Handle("/v1/order/", nex.Handler(createOrder)).Methods("GET")                  //创建订单
	router.Handle("/v1/order/notify/wechat", nex.Handler(wechatCallback)).Methods("POST") //微信订单回调
	return router
}

func CreateOrder(r *protocol.CreateOrderRequest) (interface{}, error) {
	order := &model.Order{
		OrderId:      strings.Replace(uuid.New(), "-", "", -1),
		AppId:        r.AppID,
		Uid:          r.Uid,
		ChannelId:    r.ChannelID,
		PayPlatform:  r.Platform,
		Extra:        r.Extra,
		ProductName:  r.ProductionName,
		ProductCount: r.ProductCount,
		CreatedAt:    time.Now().Unix(),
		Status:       db.OrderStatusCreated,
		Remote:       r.Device.Remote,
		Ip:           r.Device.Remote,
		Imei:         r.Device.IMEI,
		Model:        r.Device.Model,
		Os:           r.Device.OS,
	}

	resp, err := provider2.Wechat.CreateOrderResponse(order)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	if err := db.InsertOrder(order); err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	return resp, nil
}

func createOrder(r *http.Request) (interface{}, error) {
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
		return nil, errutil.ErrIllegalParameter
	}

	request := &protocol.CreateOrderRequest{
		AppID:          appId,
		ChannelID:      channelId,
		Platform:       platform,
		ProductionName: name,
		ProductCount:   count,
		Extra:          extra,
		Uid:            uid,
		Device:         protocol.Device{Remote: r.RemoteAddr},
	}

	return CreateOrder(request)
}

func TradeList(r *protocol.TradeListRequest) ([]protocol.TradeInfo, int, error) {
	if r == nil {
		return nil, 0, errutil.ErrIllegalParameter
	}

	list, total, err := db.TradeList(
		r.AppID,
		r.ChannelID,
		r.OrderID,
		r.Start,
		r.End,
		r.Offset,
		r.Count)

	if err != nil {
		return nil, 0, err
	}

	result := make([]protocol.TradeInfo, len(list))
	for i, order := range list {
		result[i] = protocol.TradeInfo{
			OrderId:        order.OrderId,
			Uid:            strconv.FormatInt(order.Uid, 10),
			Money:          order.Money,
			RealMoney:      order.RealMoney,
			ProductName:    order.ProductName,
			ProductCount:   order.ProductCount,
			ServerName:     order.ServerName,
			RoleId:         order.RoleId,
			PayBy:          order.PayPlatform,
			AppId:          order.AppId,
			ChannelId:      order.ChannelId,
			PayPlatformUid: order.ComsumerId,
			PayAt:          order.PayAt,
			Currency:       order.Currency,
		}

	}
	return result, total, nil
}

func OrderList(r *protocol.OrderListRequest) ([]protocol.OrderInfo, int, error) {
	if r == nil {
		return nil, 0, errutil.ErrIllegalParameter
	}

	id, err := strconv.ParseInt(r.Uid, 10, 0)
	if err != nil {
		return nil, 0, err
	}

	list, total, err := db.OrderList(
		id,
		r.AppID,
		r.ChannelID,
		r.OrderID,
		r.PayBy,
		r.Start,
		r.End,
		int(r.Status),
		r.Offset,
		r.Count)

	if err != nil {
		return nil, 0, err
	}

	result := make([]protocol.OrderInfo, len(list))
	for i, order := range list {
		result[i] = protocol.OrderInfo{
			OrderId:      order.OrderId,
			Uid:          strconv.FormatInt(order.Uid, 10),
			Money:        order.Money,
			RealMoney:    order.RealMoney,
			ProductName:  order.ProductName,
			ProductCount: order.ProductCount,
			ServerName:   order.ServerName,
			RoleID:       order.RoleId,
			PayBy:        order.PayPlatform,
			AppId:        order.AppId,
			Imei:         order.Imei,
			Status:       order.Status,
			Extra:        order.Extra,
			CreatedAt:    order.CreatedAt,
		}

	}
	return result, total, nil
}

//订单列表
func orderList(r *http.Request, form *nex.Form) (*protocol.OrderListResponse, error) {
	if !whitelist.VerifyIP(r.RemoteAddr) {
		return nil, errutil.ErrPermissionDenied
	}

	request := &protocol.OrderListRequest{}
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	request.Offset = form.IntOrDefault("offset", 0)
	request.Count = form.IntOrDefault("count", -1)
	request.PayBy = strings.ToLower(form.Get("pay_by"))
	request.Status = uint8(form.IntOrDefault(form.Get("status"), 0))
	request.AppID = form.Get("appid")
	request.ChannelID = form.Get("channel_id")
	request.Start = form.Int64OrDefault(form.Get("start"), -1)
	request.End = form.Int64OrDefault(form.Get("end"), -1)
	request.Uid = form.Get("uid")
	request.OrderID = form.Get("order_id")

	list, total, err := OrderList(request)
	if err != nil {
		return nil, err
	}
	return &protocol.OrderListResponse{Data: list, Total: total}, nil
}

func WechatNotify(r *protocol.WechatOrderCallbackRequest) (resp interface{}, err error) {
	var trade *model.Trade
	var order *model.Order
	if trade, resp, err = provider2.Wechat.Notify(r); err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	if order, err = db.QueryOrder(trade.OrderId); err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	if err := db.InsertTrade(trade); err != nil {
		//如果是重复通知,直接忽略之
		if err == errutil.ErrTradeExisted {
			return resp, nil
		}

		logger.Error(err.Error())
		return nil, err
	}

	if err := db.UserAddCoin(order.Uid, int64(10)); err != nil {
		logger.Error(err.Error())
		return nil, err
	}
	return resp, nil
}

func wechatCallback(_ context.Context, r *http.Request) (interface{}, error) {
	request := &protocol.WechatOrderCallbackRequest{}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if err := xml.Unmarshal(data, request); err != nil {
		return nil, err
	}

	return WechatNotify(request)
}
