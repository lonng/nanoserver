package protocol

type RegisterUserRequest struct {
	Type       int    `json:"type"`        //注册方式: 1-手机 2-贪玩蛇
	Name       string `json:"name"`        //用户名, 可空,当非游客注册时用户名与手机号必须至少出现一项
	Password   string `json:"password"`    //MD5后的用户密码, 长度>=6
	VerifyID   string `json:"verify_id"`   //验证码ID
	VerifyCode string `json:"verify_code"` //验证码
	Phone      string `json:"phone"`       //手机号,可空
	AppID      string `json:"appid"`       //来自于哪一个应用的注册
	ChannelID  string `json:"channel_id"`  //来自于哪一个渠道的注册
	Device     Device `json:"device"`      //设备信息
	Token      string `json:"token"`       //Token, 游客注册并绑定时, 验证游客身份
}

type CheckUserInfoRequest struct {
	Type       int    `json:"type"`        //注册方式: 1-手机 2-贪玩蛇
	Name       string `json:"name"`        //用户名, 可空,当非游客注册时用户名与手机号必须至少出现一项
	VerifyID   string `json:"verify_id"`   //验证码ID
	VerifyCode string `json:"verify_code"` //验证码
	Phone      string `json:"phone"`       //手机号,可空
	AppID      string `json:"appid"`
}

type UserListRequest struct {
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

type UserListResponse struct {
	Code  int        `json:"code"`  //状态码
	Data  []UserInfo `json:"users"` //用户列表
	Total int        `json:"total"`
}

type UserInfoRequest struct {
	UID int64 `json:"uid"`
}

type UserInfoResponse struct {
	Code int      `json:"code"` //状态码
	Data UserInfo `json:"data"` //数据
}

type DeleteUserRequest struct {
	UID int64 `json:"uid"`
}

type QueryUserRequest struct {
	Name string `json:"name"` //用户名
}

//用属性查询用户
type QueryUserByAttrRequest struct {
	Attr string `json:"attr"` //属性
}

//用户统计信息列表
type UserStatsInfoListRequest struct {
	RoleTypes []uint8 `json:"role_types"`
	Account   string  `json:"account"`
}

type UserStatsInfoListResponse struct {
	Code  int        `json:"code"`  //状态码
	Data  []UserInfo `json:"users"` //用户列表
	Total int        `json:"total"`
}

type QueryInfo struct {
	Name        string `json:"name"`
	MaskedPhone string `json:"masked_phone"`
}

type QueryUserResponse struct {
	Code int       `json:"code"`
	Data QueryInfo `json:"data"`
}
