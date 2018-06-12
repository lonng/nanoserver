package model

type AdminRecharge struct {
	Id           int64  `xorm:"BIGINT(20)"`
	AgentId      int64  `xorm:"not null BIGINT(20)"`
	AgentName    string `xorm:"not null VARCHAR(32)"`
	AgentAccount string `xorm:"not null VARCHAR(32)"`
	AdminId      string `xorm:"not null VARCHAR(32)"`
	AdminName    string `xorm:"not null VARCHAR(32)"`
	AdminAccount string `xorm:"not null VARCHAR(32)"`
	Extra        string `xorm:"not null VARCHAR(255)"`
	CreateAt     int64  `xorm:"not null BIGINT(20)"`
	CardCount    int64  `xorm:"not null BIGINT(20)"`
}

type Agent struct {
	Id             int64  `xorm:"BIGINT(20)"`
	Name           string `xorm:"not null VARCHAR(32)"`
	Account        string `xorm:"not null VARCHAR(32)"`
	Password       string `xorm:"not null VARCHAR(64)"`
	Phone          string `xorm:"not null VARCHAR(11)"`
	Wechat         string `xorm:"not null VARCHAR(32)"`
	Salt           string `xorm:"not null VARCHAR(32)"`
	Role           int    `xorm:"not null TINYINT(4)"`
	Status         int    `xorm:"not null TINYINT(4)"`
	Extra          string `xorm:"not null VARCHAR(255)"`
	CreateAt       int64  `xorm:"not null BIGINT(20)"`
	DeleteAt       int64  `xorm:"not null BIGINT(20)"`
	DeleteAccount  string `xorm:"not null VARCHAR(32)"`
	CreateAccount  string `xorm:"not null VARCHAR(32)"`
	ConfirmAccount string `xorm:"not null VARCHAR(32)"`
	CardCount      int64  `xorm:"not null BIGINT(20)"`
	Level          int    `xorm:"not null INT(20)"`
	Discount       int    `xorm:"not null INT(20)"`
}

type App struct {
	Id          int64  `xorm:"BIGINT(20)"`
	Name        string `xorm:"not null VARCHAR(255)"`
	Appid       string `xorm:"not null index VARCHAR(32)"`
	CpId        string `xorm:"not null VARCHAR(32)"`
	AppKey      string `xorm:"not null VARCHAR(128)"`
	AppSecret   string `xorm:"not null VARCHAR(512)"`
	RedirectUrl string `xorm:"VARCHAR(255)"`
	Extra       string `xorm:"not null VARCHAR(1024)"`
	Status      int    `xorm:"not null default 1 TINYINT(10)"`
}

type CardConsume struct {
	Id        int64  `xorm:"BIGINT(20)"`
	UserId    int64  `xorm:"not null index BIGINT(20)"`
	CardCount int    `xorm:"not null TINYINT(4)"`
	DeskId    int64  `xorm:"not null BIGINT(20)"`
	ClubId    int64  `xorm:"not null index BIGINT(20)"`
	DeskNo    string `xorm:"not null VARCHAR(32)"`
	ConsumeAt int64  `xorm:"not null BIGINT(20)"`
	Extra     string `xorm:"not null VARCHAR(255)"`
}

type Desk struct {
	Id           int64  `xorm:"BIGINT(20)"`
	Creator      int64  `xorm:"not null index BIGINT(20)"`
	ClubId       int64  `xorm:"not null index BIGINT(20)"`
	Round        int    `xorm:"not null default 8 INT(11)"`
	Mode         int    `xorm:"not null default 3 INT(11)"`
	DeskNo       string `xorm:"not null index default '' VARCHAR(6)"`
	Player0      int64  `xorm:"not null index default 0 BIGINT(20)"`
	Player1      int64  `xorm:"not null index default 0 BIGINT(20)"`
	Player2      int64  `xorm:"not null index default 0 BIGINT(20)"`
	Player3      int64  `xorm:"not null index default 0 BIGINT(20)"`
	PlayerName0  string `xorm:"not null default '' VARCHAR(255)"`
	PlayerName1  string `xorm:"not null default '' VARCHAR(255)"`
	PlayerName2  string `xorm:"not null default '' VARCHAR(255)"`
	PlayerName3  string `xorm:"not null default '' VARCHAR(255)"`
	ScoreChange0 int    `xorm:"default 0 INT(255)"`
	ScoreChange1 int    `xorm:"default 0 INT(255)"`
	ScoreChange2 int    `xorm:"default 0 INT(255)"`
	ScoreChange3 int    `xorm:"default 0 INT(255)"`
	CreatedAt    int64  `xorm:"not null index default 0 BIGINT(255)"`
	DismissAt    int64  `xorm:"not null default 0 BIGINT(255)"`
	Extras       string `xorm:"TEXT"`
}

type History struct {
	Id           int64  `xorm:"BIGINT(20)"`
	DeskId       int64  `xorm:"not null index default 0 BIGINT(20)"`
	Mode         int    `xorm:"not null index default 3 INT(255)"`
	BeginAt      int64  `xorm:"not null default 0 BIGINT(255)"`
	EndAt        int64  `xorm:"not null default 0 BIGINT(255)"`
	PlayerName0  string `xorm:"not null default '' VARCHAR(255)"`
	PlayerName1  string `xorm:"not null default '' VARCHAR(255)"`
	PlayerName2  string `xorm:"not null default '' VARCHAR(255)"`
	PlayerName3  string `xorm:"not null default '' VARCHAR(255)"`
	ScoreChange0 int    `xorm:"not null INT(255)"`
	ScoreChange1 int    `xorm:"not null default 0 INT(255)"`
	ScoreChange2 int    `xorm:"not null INT(255)"`
	ScoreChange3 int    `xorm:"not null default 0 INT(255)"`
	Snapshot     string `xorm:"TEXT"`
}

type Login struct {
	Id        int64  `xorm:"BIGINT(20)"`
	Uid       int64  `xorm:"not null index BIGINT(20)"`
	Remote    string `xorm:"not null VARCHAR(40)"`
	Ip        string `xorm:"not null VARCHAR(40)"`
	Model     string `xorm:"VARCHAR(64)"`
	Imei      string `xorm:"VARCHAR(32)"`
	Os        string `xorm:"VARCHAR(64)"`
	AppId     string `xorm:"not null VARCHAR(64)"`
	ChannelId string `xorm:"not null VARCHAR(32)"`
	LoginAt   int64  `xorm:"not null BIGINT(11)"`
	LogoutAt  int64  `xorm:"BIGINT(11)"`
}

type Online struct {
	Id        int64 `xorm:"BIGINT(20)"`
	Time      int64 `xorm:"not null BIGINT(20)"`
	UserCount int   `xorm:"not null INT(20)"`
	DeskCount int   `xorm:"not null INT(11)"`
}

type Operation struct {
	Id         int64  `xorm:"BIGINT(20)"`
	OperatorId int64  `xorm:"not null BIGINT(20)"`
	Uid        int64  `xorm:"not null index BIGINT(20)"`
	Remote     string `xorm:"not null VARCHAR(40)"`
	Ip         string `xorm:"not null VARCHAR(40)"`
	Imei       string `xorm:"VARCHAR(20)"`
	Model      string `xorm:"VARCHAR(20)"`
	Os         string `xorm:"VARCHAR(20)"`
	AppId      string `xorm:"index VARCHAR(32)"`
	OperateAt  int64  `xorm:"not null BIGINT(11)"`
	Action     int    `xorm:"not null TINYINT(255)"`
}

type Order struct {
	Id             int64  `xorm:"BIGINT(20)"`
	OrderId        string `xorm:"not null unique VARCHAR(32)"`
	Type           int    `xorm:"not null default 0 TINYINT(1)"`
	AppId          string `xorm:"not null index VARCHAR(32)"`
	ChannelId      string `xorm:"not null index VARCHAR(32)"`
	OrderPlatform  string `xorm:"VARCHAR(32)"`
	PayPlatform    string `xorm:"VARCHAR(32)"`
	ChannelOrderId string `xorm:"VARCHAR(255)"`
	Currency       string `xorm:"VARCHAR(255)"`
	Extra          string `xorm:"VARCHAR(1024)"`
	Money          int    `xorm:"not null INT(11)"`
	RealMoney      int    `xorm:"INT(11)"`
	Uid            int64  `xorm:"not null index BIGINT(20)"`
	RoleId         string `xorm:"VARCHAR(255)"`
	RoleName       string `xorm:"VARCHAR(255)"`
	ServerId       string `xorm:"VARCHAR(255)"`
	ServerName     string `xorm:"VARCHAR(255)"`
	CreatedAt      int64  `xorm:"BIGINT(11)"`
	ProductId      string `xorm:"VARCHAR(255)"`
	ProductCount   int    `xorm:"INT(10)"`
	ProductName    string `xorm:"VARCHAR(255)"`
	ProductExtra   string `xorm:"VARCHAR(255)"`
	NotifyUrl      string `xorm:"VARCHAR(2048)"`
	Status         int    `xorm:"not null default 1 TINYINT(2)"`
	Remote         string `xorm:"not null VARCHAR(40)"`
	Ip             string `xorm:"not null VARCHAR(40)"`
	Imei           string `xorm:"VARCHAR(64)"`
	Os             string `xorm:"VARCHAR(20)"`
	Model          string `xorm:"VARCHAR(20)"`
}

type Production struct {
	Id           int64  `xorm:"BIGINT(20)"`
	ProductionId string `xorm:"not null VARCHAR(32)"`
	Name         string `xorm:"not null VARCHAR(128)"`
	Extra        string `xorm:"VARCHAR(1024)"`
	Currency     string `xorm:"default '' VARCHAR(32)"`
	Price        int    `xorm:"not null INT(10)"`
	RealPrice    int    `xorm:"not null INT(10)"`
	OnlineAt     int64  `xorm:"not null BIGINT(11)"`
	OfflineAt    int64  `xorm:"BIGINT(11)"`
	Type         int    `xorm:"not null INT(1)"`
}

type Rank struct {
	Id       int64  `xorm:"BIGINT(20)"`
	Uid      int64  `xorm:"not null BIGINT(11)"`
	DeskNo   string `xorm:"not null VARCHAR(6)"`
	ClubId   int64  `xorm:"not null default -1 index BIGINT(11)"`
	Creator  int64  `xorm:"not null default 0 index BIGINT(11)"`
	Name     string `xorm:"not null default 0 VARCHAR(128)"`
	Score    int64  `xorm:"not null default 0 BIGINT(255)"`
	Match    int64  `xorm:"not null default 0 BIGINT(255)"`
	RecordAt int64  `xorm:"not null default 0 BIGINT(255)"`
}

type Recharge struct {
	Id           int64  `xorm:"BIGINT(20)"`
	AgentId      string `xorm:"not null VARCHAR(32)"`
	AgentName    string `xorm:"not null VARCHAR(32)"`
	AgentAccount string `xorm:"not null VARCHAR(32)"`
	PlayerId     int64  `xorm:"not null BIGINT(20)"`
	Extra        string `xorm:"not null VARCHAR(255)"`
	CreateAt     int64  `xorm:"not null BIGINT(20)"`
	CardCount    int64  `xorm:"not null BIGINT(20)"`
}

type Register struct {
	Id           int64  `xorm:"BIGINT(20)"`
	Uid          int64  `xorm:"not null index BIGINT(20)"`
	Remote       string `xorm:"not null VARCHAR(40)"`
	Ip           string `xorm:"not null VARCHAR(40)"`
	Imei         string `xorm:"VARCHAR(128)"`
	Os           string `xorm:"VARCHAR(20)"`
	Model        string `xorm:"VARCHAR(20)"`
	AppId        string `xorm:"not null index VARCHAR(32)"`
	ChannelId    string `xorm:"not null index VARCHAR(32)"`
	RegisterAt   int64  `xorm:"not null index BIGINT(11)"`
	RegisterType int    `xorm:"not null index TINYINT(8)"`
}

type ThirdAccount struct {
	Id           int64  `xorm:"BIGINT(20)"`
	ThirdAccount string `xorm:"not null index VARCHAR(128)"`
	Uid          int64  `xorm:"not null BIGINT(20)"`
	Platform     string `xorm:"not null index VARCHAR(32)"`
	ThirdName    string `xorm:"not null VARCHAR(64)"`
	HeadUrl      string `xorm:"not null VARCHAR(512)"`
	Sex          int    `xorm:"not null default 0 TINYINT(4)"`
}

type ThirdProperty struct {
	Id       int64  `xorm:"BIGINT(10)"`
	AppId    string `xorm:"not null VARCHAR(32)"`
	Platform string `xorm:"not null VARCHAR(32)"`
	Key      string `xorm:"VARCHAR(255)"`
	Value    string `xorm:"VARCHAR(1024)"`
}

type Trade struct {
	Id            int64  `xorm:"BIGINT(20)"`
	OrderId       string `xorm:"not null pk default '' index VARCHAR(32)"`
	PayOrderId    string `xorm:"not null VARCHAR(255)"`
	PayPlatform   string `xorm:"not null VARCHAR(32)"`
	PayAt         int64  `xorm:"BIGINT(11)"`
	PayCreateAt   int64  `xorm:"BIGINT(11)"`
	ComsumerId    string `xorm:"VARCHAR(128)"`
	MerchantId    string `xorm:"VARCHAR(128)"`
	ComsumerEmail string `xorm:"VARCHAR(64)"`
	Raw           string `xorm:"VARCHAR(2048)"`
}

type User struct {
	Id              int64  `xorm:"BIGINT(20)"`
	Algo            string `xorm:"VARCHAR(16)"`
	Hash            string `xorm:"VARCHAR(64)"`
	Salt            string `xorm:"VARCHAR(64)"`
	Role            int    `xorm:"not null TINYINT(3)"`
	Status          int    `xorm:"not null default 1 TINYINT(3)"`
	IsOnline        int    `xorm:"not null default 1 TINYINT(1)"`
	LastLoginAt     int64  `xorm:"not null index BIGINT(11)"`
	PrivKey         string `xorm:"not null VARCHAR(512)"`
	PubKey          string `xorm:"not null VARCHAR(128)"`
	Coin            int64  `xorm:"not null default 0 BIGINT(20)"`
	RegisterAt      int64  `xorm:"not null default 0 index BIGINT(20)"`
	FirstRechargeAt int64  `xorm:"not null default 0 index BIGINT(20)"`
	Debug           int    `xorm:"not null default 0 index TINYINT(1)"`
}

type Uuid struct {
	Id        int64  `xorm:"BIGINT(20)"`
	UidInUse  int64  `xorm:"not null index BIGINT(20)"`
	UidOrigin int64  `xorm:"not null BIGINT(20)"`
	Appid     string `xorm:"not null index VARCHAR(32)"`
	Uuid      string `xorm:"not null VARCHAR(64)"`
}

type Club struct {
	Id        int64  `xorm:"BIGINT(20)"`
	Balance   int64  `xorm:"not null BIGINT(20)"`
	ClubId    int64  `xorm:"not null index BIGINT(20)"`
	AgentId   int64  `xorm:"not null index BIGINT(20)"`
	Name      string `xorm:"not null VARCHAR(128)"`
	Desc      string `xorm:"not null VARCHAR(512)"`
	Member    int    `xorm:"not null default 0 INT(11)"`
	MaxMember int    `xorm:"not null default 500 INT(11)"`
	CreatedAt int64  `xorm:"not null BIGINT(20)"`
}

type UserClub struct {
	Id        int64 `xorm:"BIGINT(20)"`
	Uid       int64 `xorm:"not null index BIGINT(20)"`
	ClubId    int64 `xorm:"not null index BIGINT(20)"`
	CreatedAt int64 `xorm:"not null BIGINT(20)"`
	Status    int   `xorm:"not null default 1 TINYINT(3)"`
}
