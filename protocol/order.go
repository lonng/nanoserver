package protocol

type CreateOrderRequest struct {
	AppID          string //来自哪个应用的订单
	ChannelID      string //来自哪个渠道的订单
	Platform       string //支付平台
	ProductionName string
	ProductCount   int    //房卡数量
	Extra          string //描述信息
	Device         Device //设备信息
	Uid            int64  //Token
}

type CreateOrderByAdminRequest struct {
	AppID     string `json:"appid"`      //来自哪个应用的订单
	ChannelID string `json:"channel_id"` //来自哪个渠道的订单
	Extra     string `json:"extra"`      //描述信息
	Operator  string `json:"operator"`   //管理员账号
	Money     int    `json:"money"`      //金额
	Uid       int64  `json:"uid"`        //用户ID
	Device    Device `json:"device"`     //设备信息
}

type BalanceListRequest struct {
	Uids []string `json:"uids"` //uid列表
}

type BalanceListResponse struct {
	Code int               `json:"code"` //状态码
	Data map[string]string `json:"data"` //渠道列表
}

//OrderByAdminListRequest 由管理员创建的订单列表
type OrderByAdminListRequest struct {
	Offset    int    `json:"offset"`
	Count     int    `json:"count"`
	Start     int64  `json:"start"` //时间起点
	End       int64  `json:"end"`   //时间终点
	Uid       int64  `json:"uid"`   //用户id
	OrderId   string `json:"order_id"`
	AppID     string `json:"appid"`      //来自哪个应用的订单
	ChannelID string `json:"channel_id"` //来自哪个渠道的订单
}

//PayOrderListRequest 由管理员创建的订单列表
type PayOrderListRequest struct {
	Offset    int    `json:"offset"`
	Count     int    `json:"count"`
	Type      int    `json:"type"`       //查询类型: 1-购买代币 2-消费代币
	Start     int64  `json:"start"`      //时间起点
	End       int64  `json:"end"`        //时间终点
	Uid       int64  `json:"uid"`        //用户id
	OrderID   string `json:"order_id"`   //订单号
	AppID     string `json:"appid"`      //来自哪个应用的订单
	ChannelID string `json:"channel_id"` //来自哪个渠道的订单
}

//OrderListRequest 订单列表
type OrderListRequest struct {
	Offset    int    `json:"offset"`
	Count     int    `json:"count"`
	Status    uint8  `json:"status"`
	Start     int64  `json:"start"` //时间起点
	End       int64  `json:"end"`   //时间终点
	PayBy     string `json:"pay_by"`
	Uid       string `json:"uid"`        //用户id
	OrderID   string `json:"order_id"`   //订单号
	AppID     string `json:"appid"`      //来自哪个应用的订单
	ChannelID string `json:"channel_id"` //来自哪个渠道的订单
}

//TradeListRequest 交易列表
type TradeListRequest struct {
	Offset    int    `json:"offset"`
	Count     int    `json:"count"`
	Start     int64  `json:"start"`      //时间起点
	End       int64  `json:"end"`        //时间终点
	OrderID   string `json:"order_id"`   //订单号
	AppID     string `json:"appid"`      //来自哪个应用的订单
	ChannelID string `json:"channel_id"` //来自哪个渠道的订单
}

type PayOrderListResponse struct {
	Code  int                 `json:"code"`  //状态码
	Data  []SnakePayOrderInfo `json:"data"`  //渠道列表
	Total int                 `json:"total"` //总数量
}

type OrderListResponse struct {
	Code  int         `json:"code"`  //状态码
	Data  []OrderInfo `json:"data"`  //渠道列表
	Total int         `json:"total"` //总数量

}

type TradeListResponse struct {
	Code  int         `json:"code"`  //状态码
	Data  []TradeInfo `json:"data"`  //渠道列表
	Total int         `json:"total"` //总数量

}

type ObtainBalanceReqeust struct {
	Token string `json:"token"`
}

type ObtainBalanceResponse struct {
	Code int   `json:"code"`
	Data int64 `json:"data"`
}

type CreateOrderSnakeResponse struct {
	Result      string `json:"result"`
	PayPlatform string `json:"pay_platform"`
}

type CreateOrderWechatReponse struct {
	AppID     string `json:"appid"`
	PartnerId string `json:"partnerid"`
	OrderId   string `json:"orderid"`
	PrePayID  string `json:"prepayid"`
	NonceStr  string `json:"noncestr"`
	Sign      string `json:"sign"`
	Timestamp string `json:"timestamp"`
	Extra     string `json:"extData"`
}

type UnifyOrderCallbackRequest struct {
	PayPlatform string
	RawRequest  interface{}
}

type WechatOrderCallbackRequest struct {
	ReturnMsg        string `xml:"return_msg,omitempty"`
	DeviceInfo       string `xml:"device_info,omitempty"`
	ErrCode          string `xml:"err_code,omitempty"`
	ErrCodeDes       string `xml:"err_code_des,omitempty"`
	Attach           string `xml:"attach,omitempty"`
	CashFeeType      string `xml:"cash_fee_type,omitempty"`
	CouponFee        int    `xml:"coupon_fee,omitempty"`
	CouponCount      int    `xml:"coupon_count,omitempty"`
	CouponIDDollarN  string `xml:"coupon_id_$n,omitempty"`
	CouponFeeDollarN string `xml:"coupon_fee_$n,omitempty"`

	ReturnCode    string `xml:"return_code"`
	Appid         string `xml:"appid"`
	MchID         string `xml:"mch_id"`
	Nonce         string `xml:"nonce_str"`
	Sign          string `xml:"sign"`
	ResultCode    string `xml:"result_code"`
	Openid        string `xml:"openid"`
	IsSubscribe   string `xml:"is_subscribe"`
	TradeType     string `xml:"trade_type"`
	BankType      string `xml:"bank_type"`
	TotalFee      int    `xml:"total_fee"`
	FeeType       string `xml:"fee_type"`
	CashFee       int    `xml:"cash_fee"`
	TransactionID string `xml:"transaction_id"`
	OutTradeNo    string `xml:"out_trade_no"`
	TimeEnd       string `xml:"time_end"`

	Raw string
}

type WechatOrderCallbackResponse struct {
	ReturnCode string `xml:"return_code,cdata"`
	ReturnMsg  string `xml:"return_msg,cdata"`
}

type RechargeRequest struct {
	Count int64 `json:"count"`
	Uid   int64 `json:"uid"`
}
