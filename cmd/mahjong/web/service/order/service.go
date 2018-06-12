package order

import (
	"strconv"
	"strings"
	"time"

	"github.com/pborman/uuid"

	"github.com/lonnng/nanoserver/db"
	"github.com/lonnng/nanoserver/internal/algoutil"
	"github.com/lonnng/nanoserver/internal/errutil"
	"github.com/lonnng/nanoserver/internal/protocol"

	log "github.com/sirupsen/logrus"

	"github.com/lonnng/nanoserver/cmd/mahjong/web/provider"
	"github.com/lonnng/nanoserver/db/model"
	"github.com/lonnng/nanoserver/internal/types"
)

var logger *log.Entry

type Provider interface {
	CreateOrderResponse(*model.Order) (interface{}, error)
	Notify(request interface{}) (*model.Trade, interface{}, error)
}

type Service interface {
	CreateOrder(r *protocol.CreateOrderRequest) (interface{}, error)                           //创建订单
	YXPayOrderList(r *protocol.PayOrderListRequest) ([]protocol.SnakePayOrderInfo, int, error) //pay订单的收支列表
	OrderList(r *protocol.OrderListRequest) ([]protocol.OrderInfo, int, error)                 //获取订单的收支列表
	TradeList(r *protocol.TradeListRequest) ([]protocol.TradeInfo, int, error)                 //获取交易列表

	Notify(platform string, r interface{}) (resp interface{}, err error)

	BalanceList(uids []string) (map[string]string, error)
	GetOptions() string
}

type service struct {
}

var supportOptions = `{
	"POST": "/v1/order/",
	"POST": "/v1/order/callback/alipay",
	"POST": "/v1/order/callback/wechat",
	"POST": "/v1/order/callback/unionpay"
}`

//NewService new a service for user
func NewService(l *log.Entry) Service {
	logger = l.WithField("service", "order")
	return &service{}
}

func (s *service) CreateOrder(r *protocol.CreateOrderRequest) (interface{}, error) {
	// todo: query user by uid
	p := provider.WithName(types.YX)
	if p == nil {
		logger.Error(errutil.YXErrInvalidPayPlatform.Error())
		return nil, errutil.YXErrInvalidPayPlatform
	}
	order := &model.Order{
		OrderId:       strings.Replace(uuid.New(), "-", "", -1),
		AppId:         r.AppID,
		Uid:           r.Uid,
		ChannelId:     r.ChannelID,
		PayPlatform:   r.Platform,
		OrderPlatform: types.YX,
		Extra:         r.Extra,
		ProductName:   r.ProductionName,
		ProductCount:  r.ProductCount,
		CreatedAt:     time.Now().Unix(),
		Status:        db.OrderStatusCreated,
		Remote:        r.Device.Remote,
		Ip:            r.Device.Remote,
		Imei:          r.Device.IMEI,
		Model:         r.Device.Model,
		Os:            r.Device.OS,
	}

	resp, err := p.CreateOrderResponse(order)
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

func (s *service) BalanceList(uids []string) (map[string]string, error) {
	if uids == nil {
		return nil, errutil.YXErrIllegalParameter
	}

	return db.BalanceList(uids)

}

func (s *service) YXPayOrderList(r *protocol.PayOrderListRequest) ([]protocol.SnakePayOrderInfo, int, error) {
	if r == nil || r.Type > db.OrderTypeTest || r.Type < db.OrderTypeUnknown {
		return nil, 0, errutil.YXErrIllegalParameter
	}

	list, total, err := db.YXPayOrderList(
		r.Uid,
		r.AppID,
		r.ChannelID,
		r.OrderID,
		r.Start,
		r.End,
		r.Type,
		r.Offset,
		r.Count)

	if err != nil {
		return nil, 0, err
	}

	result := make([]protocol.SnakePayOrderInfo, len(list))
	for i, order := range list {
		result[i] = protocol.SnakePayOrderInfo{
			OrderId:      order.OrderId,
			Uid:          strconv.FormatInt(order.Uid, 10),
			Money:        order.Money,
			RealMoney:    order.RealMoney,
			ProductName:  order.ProductName,
			ProductCount: order.ProductCount,
			ServerName:   order.ServerName,
			RoleID:       order.RoleId,
			Type:         order.Type,
			AppId:        order.AppId,
			ChannelId:    order.ChannelId,
			Imei:         order.Imei,
			Status:       order.Status,
			Extra:        order.Extra,
			CreatedAt:    order.CreatedAt,
		}
	}
	return result, total, nil
}

func (s *service) OrderList(r *protocol.OrderListRequest) ([]protocol.OrderInfo, int, error) {
	if r == nil {
		return nil, 0, errutil.YXErrIllegalParameter
	}

	list, total, err := db.OrderList(
		algoutil.RetriveInt64OrDefault(r.Uid, 0),
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

func (s *service) TradeList(r *protocol.TradeListRequest) ([]protocol.TradeInfo, int, error) {
	if r == nil {
		return nil, 0, errutil.YXErrIllegalParameter
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

func (s *service) Notify(platform string, r interface{}) (resp interface{}, err error) {
	var (
		p     = provider.WithName(platform)
		trade *model.Trade
		order *model.Order
	)
	if p == nil {
		logger.Error("invalid pay platform: " + platform)
		return nil, errutil.YXErrInvalidPayPlatform
	}

	if trade, resp, err = p.Notify(r); err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	if order, err = db.QueryOrder(trade.OrderId); err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	if err := db.InsertTrade(trade); err != nil {
		//如果是重复通知,直接忽略之
		if err == errutil.YXErrTradeExisted {
			return resp, nil
		}

		logger.Error(err.Error())
		return nil, err
	}

	prod, err := db.QueryProduction(order.ProductId)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}
	if err := db.UserAddCoin(order.Uid, int64(prod.Price)); err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	return resp, nil
}

func (*service) GetOptions() string {
	return supportOptions
}
