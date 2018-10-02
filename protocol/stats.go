package protocol

type AppStatsRequest struct {
	AppID     string  `json:"app_id"`
	ChannelID string  `json:"channel_id"`
	Remote    string  `json:"remote"`
	Event     string  `json:"event"`
	Extra     string  `json:"extra"`
	Device    *Device `json:"device"`
}

type ChannelAndAppStatsSummaryRequest struct {
	AppIds     []string `json:"app_ids"`
	ChannelIds []string `json:"channel_ids"`
	Start      int64    `json:"start"`
	End        int64    `json:"end"`
	SortBy     byte     `json:"sort_by"`
}

type UserStatsSummaryRequest struct {
	AppIds     []string `json:"app_ids"`
	ChannelIds []string `json:"channel_ids"`
	Role       byte     `json:"role"` //账号类型
	Uid        int64    `json:"uid"`
	Start      int64    `json:"start"` //注册起始时间
	End        int64    `json:"end"`   //注册结束时间
	SortBy     byte     `json:"sort_by"`
}

type ChannelAndAPPStatsSummary struct {
	Start                int64  `json:"start"`
	AppId                string `json:"app_id"`
	ChannelId            string `json:"channel_id"`
	AppName              string `json:"app_name"`
	ChannelName          string `json:"channel_name"`
	AccountInc           int64  `json:"account_inc"`            //新增用户
	DeviceInc            int64  `json:"device_inc"`             //新增设备
	TotalRecharge        int64  `json:"total_recharge"`         //总充值
	TotalRechargeAccount int64  `json:"total_recharge_account"` //总充值人数

	PaidAccountInc       int64 `json:"paid_account_inc"`        //新增付费用户
	PaidTotalRechargeInc int64 `json:"paid_total_recharge_inc"` //新增付费总额

	RegPaidAccountInc       int64 `json:"reg_paid_account_inc"`        //新增注册并付费用户
	RegPaidTotalRechargeInc int64 `json:"reg_paid_total_recharge_inc"` //新增注册并付费总额

	//PaidAccountIncRate    float32 `json:"paid_account_inc_rate"`     //新增用户付费率
	RegPaidAccountIncRate string `json:"reg_paid_account_inc_rate"` //新增注册用户付费率
}

type UserStatsSummary struct {
	Name      string `json:"name"`       //名字
	Uid       int64  `json:"uid"`        //uid
	Role      byte   `json:"role"`       //角色
	AppID     string `json:"app_id"`     //appid
	ChannelID string `json:"channel_id"` //channel id

	OS      string `json:"os"`
	IP      string `json:"ip"`       //最后登录ip
	Device  string `json:"deivce"`   //最后登录设备
	LoginAt int64  `json:"login_at"` //最后登录时间

	RegisterAt    int64 `json:"register_at"`    //注册时间
	LoginNum      int64 `json:"login_num"`      //登录次数
	RechargeNum   int64 `json:"recharge_num"`   //充值次数
	RechargeTotal int64 `json:"total_recharge"` //充值总金额
}

type ChannelAndAPPStatsSummaryResponse struct {
	Code  int                          `json:"code"`  //状态码
	Data  []*ChannelAndAPPStatsSummary `json:"data"`  //
	Total int                          `json:"total"` //总数量
}

type UserStatsSummaryResponse struct {
	Code  int                 `json:"code"` //状态码
	Data  []*UserStatsSummary `json:"data"`
	Total int                 `json:"total"` //总数量
}

type RetentionLite struct {
	Login int64  `json:"login"`
	Rate  string `json:"rate"`
}
type Retention struct {
	Date     int   `json:"date"`
	Register int64 `json:"register"`

	Retention_1  RetentionLite `json:"retention_1"`  //次日
	Retention_2  RetentionLite `json:"retention_2"`  //2日
	Retention_3  RetentionLite `json:"retention_3"`  //3日
	Retention_7  RetentionLite `json:"retention_7"`  //7日
	Retention_14 RetentionLite `json:"retention_14"` //14日
	Retention_30 RetentionLite `json:"retention_30"` //30日

}

type RetentionResponse struct {
	Code int `json:"code"`

	Data interface{} `json:"data"`
}

type RetentionListRequest struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type Rank struct {
	Uid   int64  `json:"uid"`
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type CommonStatsItem struct {
	Date  int64 `json:"date"`
	Value int64 `json:"value"`
}

//房卡消耗
type CardConsume CommonStatsItem

//活跃用户
type ActivationUser CommonStatsItem
