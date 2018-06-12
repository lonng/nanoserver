package db

//trade & order => views
type ViewTrade struct {
	PayAt int64
	Uid   int64
	Id    int64

	Type         int
	Money        int
	RealMoney    int
	ProductCount int
	Status       int

	OrderId        string
	ComsumerId     string
	AppId          string
	ChannelId      string
	OrderPlatform  string
	ChannelOrderId string
	Currency       string
	RoleId         string
	ServerName     string
	ProductId      string
	ProductName    string
	RoleName       string
	PayPlatform    string
}

func (v *ViewTrade) TableName() string {
	return "view_trade"
}

// trade & register & user => views
type ViewChannelApp struct {
	Id     int64
	Type   byte
	Status byte

	Uid             int64
	CreatedAt       int64
	RegisterAt      int64
	FirstRechargeAt int64

	RealMoney    int
	RegisterType int

	Os                string
	Imei              string
	OrderId           string
	AppId             string
	Model             string
	ServerId          string
	ProductId         string
	OrderPlatform     string
	PayPlatform       string
	PassportChannelId string
	PaymentChannelId  string
}

func (v *ViewChannelApp) TableName() string {
	return "view_channel_app"
}
