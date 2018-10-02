package protocol

type ThirdUserLoginRequest struct {
	Platform    string `json:"platform"`     //三方平台/渠道
	AppID       string `json:"appId"`        //用户来自于哪一个应用
	ChannelID   string `json:"channelId"`    //用户来自于哪一个渠道
	Device      Device `json:"device"`       //设备信息
	Name        string `json:"name"`         //微信平台名
	OpenID      string `json:"openid"`       //微信平台openid
	AccessToken string `json:"access_token"` //微信AccessToken
}

type LoginInfo struct {
	// 三方登录字段
	Platform     string `json:"platform"`      //三方平台
	ThirdAccount string `json:"third_account"` //三方平台唯一ID
	ThirdName    string `json:"account"`       //三方平台账号名

	Token      string `json:"token"`       //用户Token
	ExpireTime int64  `json:"expire_time"` //Token过期时间

	AccountID int64 `json:"acId"` //用户的uuid,即user表的pk

	GameServerIP   string `json:"ip"` //游戏服的ip&port
	GameServerPort int    `json:"port"`
}

type UserLoginResponse struct {
	Code int32     `json:"code"` //状态码
	Data LoginInfo `json:"data"`
}

type LoginRequest struct {
	AppID     string `json:"appId"`     //用户来自于哪一个应用
	ChannelID string `json:"channelId"` //用户来自于哪一个渠道
	IMEI      string `json:"imei"`
	Device    Device `json:"device"`
}

type ClientConfig struct {
	Version     string `json:"version"`
	Android     string `json:"android"`
	IOS         string `json:"ios"`
	Heartbeat   int    `json:"heartbeat"`
	ForceUpdate bool   `json:"forceUpdate"`

	Title string `json:"title"` // 分享标题
	Desc  string `json:"desc"`  // 分享描述

	Daili1 string `json:"daili1"`
	Daili2 string `json:"daili2"`
	Kefu1  string `json:"kefu1"`

	AppId  string `json:"appId"`
	AppKey string `json:"appKey"`
}

type LoginResponse struct {
	Code     int          `json:"code"`
	Name     string       `json:"name"`
	Uid      int64        `json:"uid"`
	HeadUrl  string       `json:"headUrl"`
	FangKa   int64        `json:"fangka"`
	Sex      int          `json:"sex"` //[0]未知 [1]男 [2]女
	IP       string       `json:"ip"`
	Port     int          `json:"port"`
	PlayerIP string       `json:"playerIp"`
	Config   ClientConfig `json:"config"`
	Messages []string     `json:"messages"`
	ClubList []ClubItem   `json:"clubList"`
	Debug    int          `json:"debug"`
}

type LoginToGameServerResponse struct {
	Uid      int64  `json:"acId"`
	Nickname string `json:"nickname"`
	HeadUrl  string `json:"headURL"`
	Sex      int    `json:"sex"`
	FangKa   int    `json:"fangka"`
}

type LoginToGameServerRequest struct {
	Name    string `json:"name"`
	Uid     int64  `json:"uid"`
	HeadUrl string `json:"headUrl"`
	Sex     int    `json:"sex"` //[0]未知 [1]男 [2]女
	FangKa  int    `json:"fangka"`
	IP      string `json:"ip"`
}

type EncryptTest struct {
	Payload string `json:"payload"`
	Key     string `json:"key"`
}

type EncryptTestTest struct {
	Result string `json:"result"`
}
