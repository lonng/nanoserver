package protocol

import "fmt"

type AppInfo struct {
	Name            string                       `json:"name"`             //应用名
	AppID           string                       `json:"appid"`            //应用id
	AppKey          string                       `json:"appkey"`           //应用key
	RedirectURI     string                       `json:"redirect_uri"`     //注册时填的redirect_uri
	Extra           string                       `json:"extra"`            //额外信息
	ThirdProperties map[string]map[string]string `json:"third_properties"` //此app在第三方平台(eg: wechat)上的相关配置
}

type SnakePayOrderInfo struct {
	OrderId      string `json:"order_id"`      //订单号
	Uid          string `json:"uid"`           //接收者id
	ServerName   string `json:"server_name"`   //区服名
	RoleID       string `json:"role_id"`       //角色id
	AppId        string `json:"appid"`         //应用id
	ChannelId    string `json:"channel_id"`    //渠道id
	Extra        string `json:"extra"`         //额外信息
	Imei         string `json:"imei"`          //imei
	ProductName  string `json:"product_name"`  //商品名
	Type         int    `json:"type"`          //收支类型: 1-购买代币 2-消费代币
	Money        int    `json:"money"`         //标价
	RealMoney    int    `json:"real_money"`    //实际售价
	ProductCount int    `json:"product_count"` //商品数量
	Status       int    `json:"status"`        // 订单状态 1-创建 2-完成 3-游戏服务器已经确认

	CreatedAt int64 `json:"created_at"` //发放时间
}

type OrderInfo struct {
	OrderId      string `json:"order_id"`      //订单号
	Uid          string `json:"uid"`           //接收者id
	AppId        string `json:"appid"`         //应用id
	ServerName   string `json:"server_name"`   //区服名
	RoleID       string `json:"role_id"`       //角色id
	Extra        string `json:"extra"`         //额外信息
	Imei         string `json:"imei"`          //imei
	ProductName  string `json:"product_name"`  //商品名
	PayBy        string `json:"pay_by"`        //收支类型: alipay, wechat ...
	ProductCount int    `json:"product_count"` //商品数量
	Money        int    `json:"money"`         //标价
	RealMoney    int    `json:"real_money"`    //实际售价
	Status       int    `json:"status"`        // 订单状态 1-创建 2-完成 3-游戏服务器已经确认
	CreatedAt    int64  `json:"created_at"`    //发放时间
}

type TradeInfo struct {
	OrderId        string `json:"order_id"`         //订单号
	Uid            string `json:"uid"`              //snake uid
	PayPlatformUid string `json:"pay_platform_uid"` //支付平台uid
	AppId          string `json:"appid"`            //应用id
	ChannelId      string `json:"channel_id"`       //渠道id
	ProductName    string `json:"product_name"`     //商品名
	PayBy          string `json:"pay_by"`           //收支类型: alipay, wechat ...
	ServerName     string `json:"server_name"`
	RoleName       string `json:"role_name"`
	RoleId         string `json:"role_id"`
	Currency       string `json:"currency"`
	ProductCount   int    `json:"product_count"` //商品数量
	Money          int    `json:"money"`         //标价
	RealMoney      int    `json:"real_money"`    //实际售价
	PayAt          int64  `json:"pay_at"`        //支付时间

}

type UserInfo struct {
	UID         int64  `json:"uid"`
	Name        string `json:"name"`            //用户名, 可空,当非游客注册时用户名与手机号必须至少出现一项
	Phone       string `json:"phone"`           //手机号,可空
	Role        int    `json:"role"`            //玩家类型
	Status      int    `json:"status"`          //状态
	IsOnline    int    `json:"is_online"`       //是否在线
	LastLoginAt int64  `json:"last_login_time"` //最后登录时间
}

type DailyStats struct {
	Score     int      `json:"score"`      //战绩
	AsCreator int64    `json:"as_creator"` //开房次数
	Win       int      `json:"win"`        // 赢的次数
	DeskNos   []string `json:"desks"`      //所参加的桌号

}

type UserStatsInfo struct {
	ID             int64  `json:"id"`
	Uid            int64  `json:"uid"`
	Name           string `json:"name"`
	RegisterAt     int64  `json:"register_at"`
	RegisterIP     string `json:"register_ip"`
	LastestLoginAt int64  `json:"lastest_login_at"`
	LastestLoginIP string `json:"lastest_login_ip"`

	TotalMatch int64 `json:"total_match"` //总对局数
	RemainCard int64 `json:"remain_card"` //剩余房卡

	StatsAt []int64               //统计时间
	Stats   map[int64]*DailyStats //时间对应的数据
}

type Device struct {
	IMEI   string `json:"imei"`   //设备的imei号
	OS     string `json:"os"`     //os版本号
	Model  string `json:"model"`  //硬件型号
	IP     string `json:"ip"`     //内网IP
	Remote string `json:"remote"` //外网IP
}

type StringResponse struct {
	Code int    `json:"code"` //状态码
	Data string `json:"data"` //字符串数据
}

type CommonResponse struct {
	Code int         `json:"code"` //状态码
	Data interface{} `json:"data"` //整数状态
}

var SuccessResponse = StringResponse{0, "success"}

const (
	RegTypeThird = 5 //三方平台添加账号
)

var EmptyMessage = &None{}

type EmptyRequest struct{}

var SuccessMessage = &StringMessage{Message: "success"}

type None struct{}

type StringMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type ErrorMessage struct {
	ErrorType int    `json:"errorType"`
	Message   string `json:"msg"`
}

type DailyMatchProgressInfo struct {
	HasProgress  bool  `json:"hasProgress"`
	IsHaveFanPai bool  `json:"isHaveFanPai"`
	Heart        int   `json:"heart"` //最大只能是3
	BaoPaiMax    int   `json:"baoPaiMax"`
	BaoPaiNum    int   `json:"baoPaiNum"`
	Coin         int64 `json:"coin"`
	Score        int   `json:"score"`
	RoomType     int   `json:"roomType"`
	BaoPaiID     int   `json:"baoPaiId"`
}

type PlayerReady struct {
	Account int64 `json:"account"`
}

//听牌信息
type Ting struct {
	Index int   `json:"index"`
	Hu    []int `json:"hu"`
}

//所有被听的牌
type Tings []Ting

//所有可执行的操作
type Ops []Op

//提示
type Hint struct {
	Ops   Ops   `json:"ops"`
	Tings Tings `json:"tings"`
	Uid   int64 `json:"uid"`
}

func (h *Hint) String() string {
	return fmt.Sprintf("UID=%d, Ops=%+v, Tings=%+v", h.Uid, h.Ops, h.Tings)
}
