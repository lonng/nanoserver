package model

type AdminRecharge struct {
	Id           int64
	AgentId      int64  `xorm:"not null BIGINT(20) default"`
	AgentName    string `xorm:"not null VARCHAR(32) default"`
	AgentAccount string `xorm:"not null VARCHAR(32) default"`
	AdminId      string `xorm:"not null VARCHAR(32) default"`
	AdminName    string `xorm:"not null VARCHAR(32) default"`
	AdminAccount string `xorm:"not null VARCHAR(32) default"`
	Extra        string `xorm:"not null VARCHAR(255) default"`
	CreateAt     int64  `xorm:"not null BIGINT(20) default"`
	CardCount    int64  `xorm:"not null BIGINT(20) default"`
}

type Agent struct {
	Id             int64
	Name           string `xorm:"not null VARCHAR(32) default"`
	Account        string `xorm:"not null VARCHAR(32) default"`
	Password       string `xorm:"not null VARCHAR(64) default"`
	Phone          string `xorm:"not null VARCHAR(11) default"`
	Wechat         string `xorm:"not null VARCHAR(32) default"`
	Salt           string `xorm:"not null VARCHAR(32) default"`
	Role           int    `xorm:"not null TINYINT(4) default"`
	Status         int    `xorm:"not null TINYINT(4) default"`
	Extra          string `xorm:"not null VARCHAR(255) default"`
	CreateAt       int64  `xorm:"not null BIGINT(20) default"`
	DeleteAt       int64  `xorm:"not null BIGINT(20) default"`
	DeleteAccount  string `xorm:"not null VARCHAR(32) default"`
	CreateAccount  string `xorm:"not null VARCHAR(32) default"`
	ConfirmAccount string `xorm:"not null VARCHAR(32) default"`
	CardCount      int64  `xorm:"not null BIGINT(20) default"`
	Level          int    `xorm:"not null INT(20) default"`
	Discount       int    `xorm:"not null INT(20) default"`
}

type CardConsume struct {
	Id        int64
	UserId    int64  `xorm:"not null index BIGINT(20) default"`
	CardCount int    `xorm:"not null TINYINT(4) default"`
	DeskId    int64  `xorm:"not null BIGINT(20) default"`
	ClubId    int64  `xorm:"not null index BIGINT(20) default"`
	DeskNo    string `xorm:"not null VARCHAR(32) default"`
	ConsumeAt int64  `xorm:"not null BIGINT(20) default"`
	Extra     string `xorm:"not null VARCHAR(255) default"`
}

type Desk struct {
	Id           int64
	Creator      int64  `xorm:"not null index BIGINT(20) default"`
	ClubId       int64  `xorm:"not null index BIGINT(20) default"`
	Round        int    `xorm:"not null INT(11) default 8"`
	Mode         int    `xorm:"not null INT(11) default 3"`
	DeskNo       string `xorm:"not null index VARCHAR(6) default"`
	Player0      int64  `xorm:"not null index BIGINT(20) default 0"`
	Player1      int64  `xorm:"not null index BIGINT(20) default 0"`
	Player2      int64  `xorm:"not null index BIGINT(20) default 0"`
	Player3      int64  `xorm:"not null index BIGINT(20) default 0"`
	PlayerName0  string `xorm:"not null VARCHAR(255) default"`
	PlayerName1  string `xorm:"not null VARCHAR(255) default"`
	PlayerName2  string `xorm:"not null VARCHAR(255) default"`
	PlayerName3  string `xorm:"not null VARCHAR(255) default"`
	ScoreChange0 int    `xorm:"not null INT(255) default 0"`
	ScoreChange1 int    `xorm:"not null INT(255) default 0"`
	ScoreChange2 int    `xorm:"not null INT(255) default 0"`
	ScoreChange3 int    `xorm:"not null INT(255) default 0"`
	CreatedAt    int64  `xorm:"not null index BIGINT(255) default 0"`
	DismissAt    int64  `xorm:"not null BIGINT(255) default 0"`
	Extras       string `xorm:"not null TEXT default"`
}

type History struct {
	Id           int64
	DeskId       int64  `xorm:"not null index BIGINT(20) default 0"`
	Mode         int    `xorm:"not null index INT(255) default 3"`
	BeginAt      int64  `xorm:"not null BIGINT(255) default 0 "`
	EndAt        int64  `xorm:"not null BIGINT(255) default 0 "`
	PlayerName0  string `xorm:"not null VARCHAR(255) default"`
	PlayerName1  string `xorm:"not null VARCHAR(255) default"`
	PlayerName2  string `xorm:"not null VARCHAR(255) default"`
	PlayerName3  string `xorm:"not null VARCHAR(255) default"`
	ScoreChange0 int    `xorm:"not null INT(255) default 0"`
	ScoreChange1 int    `xorm:"not null INT(255) default 0"`
	ScoreChange2 int    `xorm:"not null INT(255) default 0"`
	ScoreChange3 int    `xorm:"not null INT(255) default 0"`
	Snapshot     string `xorm:"not null TEXT default"`
}

type Login struct {
	Id        int64
	Uid       int64  `xorm:"not null index BIGINT(20) default"`
	Remote    string `xorm:"not null VARCHAR(40) default"`
	Ip        string `xorm:"not null VARCHAR(40) default"`
	Model     string `xorm:"not null VARCHAR(64) default"`
	Imei      string `xorm:"not null VARCHAR(32) default"`
	Os        string `xorm:"not null VARCHAR(64) default"`
	AppId     string `xorm:"not null VARCHAR(64) default"`
	ChannelId string `xorm:"not null VARCHAR(32) default"`
	LoginAt   int64  `xorm:"not null BIGINT(11) default"`
	LogoutAt  int64  `xorm:"not null BIGINT(11) default"`
}

type Online struct {
	Id        int64
	Time      int64 `xorm:"not null BIGINT(20) default"`
	UserCount int   `xorm:"not null INT(20) default"`
	DeskCount int   `xorm:"not null INT(11) default"`
}

type Order struct {
	Id             int64
	OrderId        string `xorm:"not null unique VARCHAR(32)"`
	Type           int    `xorm:"not null TINYINT(1) default 0"`
	AppId          string `xorm:"not null index VARCHAR(32) default"`
	ChannelId      string `xorm:"not null index VARCHAR(32) default"`
	PayPlatform    string `xorm:"not null VARCHAR(32) default"`
	ChannelOrderId string `xorm:"not null VARCHAR(255) default"`
	Currency       string `xorm:"not null VARCHAR(255) default"`
	Extra          string `xorm:"not null VARCHAR(1024) default"`
	Money          int    `xorm:"not null INT(11) default"`
	RealMoney      int    `xorm:"not null INT(11) default"`
	Uid            int64  `xorm:"not null index BIGINT(20) default"`
	RoleId         string `xorm:"not null VARCHAR(255) default"`
	RoleName       string `xorm:"not null VARCHAR(255) default"`
	ServerId       string `xorm:"not null VARCHAR(255) default"`
	ServerName     string `xorm:"not null VARCHAR(255) default"`
	CreatedAt      int64  `xorm:"not null BIGINT(11) default"`
	ProductId      string `xorm:"not null VARCHAR(255) default"`
	ProductCount   int    `xorm:"not null INT(10) default"`
	ProductName    string `xorm:"not null VARCHAR(255) default"`
	ProductExtra   string `xorm:"not null VARCHAR(255) default"`
	NotifyUrl      string `xorm:"not null VARCHAR(2048) default"`
	Status         int    `xorm:"not null TINYINT(2) default 1"`
	Remote         string `xorm:"not null VARCHAR(40) default"`
	Ip             string `xorm:"not null VARCHAR(40) default"`
	Imei           string `xorm:"not null VARCHAR(64) default"`
	Os             string `xorm:"not null VARCHAR(20) default"`
	Model          string `xorm:"not null VARCHAR(20) default"`
}

type Recharge struct {
	Id           int64
	AgentId      string `xorm:"not null VARCHAR(32) default"`
	AgentName    string `xorm:"not null VARCHAR(32) default"`
	AgentAccount string `xorm:"not null VARCHAR(32) default"`
	PlayerId     int64  `xorm:"not null BIGINT(20) default"`
	Extra        string `xorm:"not null VARCHAR(255) default"`
	CreateAt     int64  `xorm:"not null BIGINT(20) default"`
	CardCount    int64  `xorm:"not null BIGINT(20) default"`
}

type Register struct {
	Id           int64
	Uid          int64  `xorm:"not null index BIGINT(20) default"`
	Remote       string `xorm:"not null VARCHAR(40) default"`
	Ip           string `xorm:"not null VARCHAR(40) default"`
	Imei         string `xorm:"not null VARCHAR(128) default"`
	Os           string `xorm:"not null VARCHAR(20) default"`
	Model        string `xorm:"not null VARCHAR(20) default"`
	AppId        string `xorm:"not null index VARCHAR(32) default"`
	ChannelId    string `xorm:"not null index VARCHAR(32) default"`
	RegisterAt   int64  `xorm:"not null index BIGINT(11) default"`
	RegisterType int    `xorm:"not null index TINYINT(8) default"`
}

type ThirdAccount struct {
	Id           int64
	ThirdAccount string `xorm:"not null index VARCHAR(128) default"`
	Uid          int64  `xorm:"not null BIGINT(20) default"`
	Platform     string `xorm:"not null index VARCHAR(32) default"`
	ThirdName    string `xorm:"not null VARCHAR(64) default"`
	HeadUrl      string `xorm:"not null VARCHAR(512) default"`
	Sex          int    `xorm:"not null TINYINT(4)  default 0"`
}

type Trade struct {
	Id            int64
	OrderId       string `xorm:"not null unique VARCHAR(32) default"`
	PayOrderId    string `xorm:"not null VARCHAR(255) default"`
	PayPlatform   string `xorm:"not null VARCHAR(32) default"`
	PayAt         int64  `xorm:"not null BIGINT(11) default"`
	PayCreateAt   int64  `xorm:"not null BIGINT(11) default"`
	ComsumerId    string `xorm:"not null VARCHAR(128) default"`
	MerchantId    string `xorm:"not null VARCHAR(128) default"`
	ComsumerEmail string `xorm:"not null VARCHAR(64) default"`
	Raw           string `xorm:"not null VARCHAR(2048) default"`
}

type User struct {
	Id              int64
	Algo            string `xorm:"not null VARCHAR(16) default"`
	Hash            string `xorm:"not null VARCHAR(64) default"`
	Salt            string `xorm:"not null VARCHAR(64) default"`
	Role            int    `xorm:"not null TINYINT(3) default 1"`
	Status          int    `xorm:"not null TINYINT(3) default 1"`
	IsOnline        int    `xorm:"not null TINYINT(1) default 1"`
	LastLoginAt     int64  `xorm:"not null index BIGINT(11) default"`
	PrivKey         string `xorm:"not null VARCHAR(512) default"`
	PubKey          string `xorm:"not null VARCHAR(128) default"`
	Coin            int64  `xorm:"not null BIGINT(20) default 0"`
	RegisterAt      int64  `xorm:"not null index BIGINT(20) default 0"`
	FirstRechargeAt int64  `xorm:"not null index BIGINT(20) default 0"`
	Debug           int    `xorm:"not null index TINYINT(1) default 0"`
}

type Uuid struct {
	Id        int64
	UidInUse  int64  `xorm:"not null index BIGINT(20) default 0"`
	UidOrigin int64  `xorm:"not null BIGINT(20) default 0"`
	Appid     string `xorm:"not null index VARCHAR(32) default"`
	Uuid      string `xorm:"not null VARCHAR(64) default"`
}

type Club struct {
	Id        int64
	Balance   int64  `xorm:"not null BIGINT(20) default 0"`
	ClubId    int64  `xorm:"not null index BIGINT(20) default 0"`
	AgentId   int64  `xorm:"not null index BIGINT(20) default 0"`
	Name      string `xorm:"not null VARCHAR(128) default"`
	Desc      string `xorm:"not null VARCHAR(512) default"`
	Member    int    `xorm:"not null INT(11) default"`
	MaxMember int    `xorm:"not null INT(11) default 500"`
	CreatedAt int64  `xorm:"not null BIGINT(20) default"`
}

type UserClub struct {
	Id        int64
	Uid       int64 `xorm:"not null index BIGINT(20) default"`
	ClubId    int64 `xorm:"not null index BIGINT(20) default"`
	CreatedAt int64 `xorm:"not null BIGINT(20) default"`
	Status    int   `xorm:"not null TINYINT(3) default 1"`
}
