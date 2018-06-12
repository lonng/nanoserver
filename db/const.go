package db

const (
	KWX = "broker"
)

const (
	defaultMaxConns = 10
)

// Users表中role字段的取值
const (
	RoleTypeAdmin = 1 //管理员账号
	RoleTypeThird = 2 //三方平台账号
)

const (
	OpActionRegister  = 1 //注册
	OpActionFreezen   = 2 //冻结账号
	OpActionUnFreezen = 3 //账号解冻
	OpActionDelete    = 4 //账号删除
)

const (
	UserOffline = 1 //离线
	UserOnline  = 2 //在线
)

const (
	StatusNormal  = 1 //正常
	StatusDeleted = 2 //删除
	StatusFreezed = 3 //冻结
	StatusBound   = 4 //绑定
)

// 订单状态
const (
	OrderStatusCreated  = 1 //创建
	OrderStatusPayed    = 2 //完成
	OrderStatusNotified = 3 //已确认订单
)

const (
	OrderTypeUnknown      = iota
	OrderTypeBuyToken     //购买令牌
	OrderTypeConsumeToken //消费代币(eg:使用令牌购买游戏中的道具,比如房卡)
	OrderTypeConsume3rd   //第三方支付平台消费(eg:直接使用alipay, wechat等购买游戏中的道具)
	OrderTypeTest         //支付测试
)

const (
	NotifyResultSuccess = 1 //通知成功
	NotifyResultFailed  = 2 //通知失败
)

const (
	dayInSecond = 24 * 60 * 60

	day1  = dayInSecond
	day2  = day1 * 2
	day3  = day1 * 3
	day7  = day1 * 7
	day14 = day1 * 14
	day30 = day1 * 30
)

const (
	RankingNormal = 1
	RankingDesc   = 2
)

const (
	DefaultTopN = 10
)
