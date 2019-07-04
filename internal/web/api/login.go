package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/lonng/nanoserver/db"
	"github.com/lonng/nanoserver/pkg/errutil"
	"github.com/lonng/nanoserver/protocol"

	"github.com/gorilla/mux"
	"github.com/lonng/nanoserver/db/model"
	"github.com/lonng/nex"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/chanxuehong/wechat.v2/open/oauth2"
)

var (
	host     string                // 服务器地址
	port     int                   // 服务器端口
	config   protocol.ClientConfig // 远程配置
	messages []string              // 广播消息
	logger   = log.WithFields(log.Fields{"component": "http", "service": "login"})

	// 游客登陆
	enableGuest   = false
	guestChannels = []string{}

	enableDebug = false
)

const defaultCoin = 10

func AddMessage(message string) {
	messages = append(messages, message)
}

func MakeLoginService() http.Handler {
	host = viper.GetString("game-server.host")
	port = viper.GetInt("game-server.port")

	// 更新相关配置
	config.Version = viper.GetString("update.version")
	config.Android = viper.GetString("update.android")
	config.IOS = viper.GetString("update.ios")
	config.Heartbeat = viper.GetInt("core.heartbeat")

	// 分享相关配置
	config.Title = viper.GetString("share.title")
	config.Desc = viper.GetString("share.desc")

	// 客服相关配置
	config.Daili1 = viper.GetString("contact.daili1")
	config.Daili2 = viper.GetString("contact.daili2")
	config.Kefu1 = viper.GetString("contact.kefu1")

	// 游客相关配置
	enableGuest = viper.GetBool("login.guest")
	guestChannels = viper.GetStringSlice("login.lists")
	logger.Infof("是否开启游客登陆: %t, 渠道列表: %v", enableGuest, guestChannels)

	// 语音相关配置
	config.AppId = viper.GetString("voice.appid")
	config.AppKey = viper.GetString("voice.appkey")

	if config.Heartbeat < 5 {
		config.Heartbeat = 5
	}

	messages = viper.GetStringSlice("broadcast.message")

	logger.Debugf("version infomation: %+v", config)
	logger.Debugf("广播消息: %v", messages)

	fu := viper.GetBool("update.force")
	logger.Infof("是否强制更新: %t", fu)
	config.ForceUpdate = fu

	router := mux.NewRouter()
	router.Handle("/v1/user/login/query", nex.Handler(queryHandler)).Methods("POST")        //三方登录
	router.Handle("/v1/user/login/3rd", nex.Handler(thirdUserLoginHandler)).Methods("POST") //三方登录
	router.Handle("/v1/user/login/guest", nex.Handler(guestLoginHandler)).Methods("POST")   //三方登录
	router.Handle("/v1/user/club", nex.Handler(clubListHandler)).Methods("GET")             // 获取俱乐部列表
	return router
}

func ip(addr string) string {
	addr = strings.TrimSpace(addr)
	deflt := "127.0.0.1"
	if addr == "" {
		return deflt
	}

	if parts := strings.Split(addr, ":"); len(parts) > 0 {
		return parts[0]
	}

	return deflt
}

// 过滤emoji表情, 设置为*
func filterEmoji(content string) string {
	new_content := ""
	for _, value := range content {
		_, size := utf8.DecodeRuneInString(string(value))
		if size <= 3 {
			new_content += string(value)
		} else {
			new_content += "*"
		}
	}
	return new_content
}

func checkSession(uid int64) {
	// fixed: 之前已有session未断开
	// 检查是否该玩家是否有未断开的网络连接, 把之前的号顶掉
	// game.Kick(uid)
}

func clubs(uid int64) []protocol.ClubItem {
	list, err := db.ClubList(uid)
	if err != nil {
		return []protocol.ClubItem{}
	}

	ret := make([]protocol.ClubItem, len(list))
	for i := range list {
		ret[i] = protocol.ClubItem{
			Id:        list[i].ClubId,
			Name:      list[i].Name,
			Desc:      list[i].Desc,
			Member:    list[i].Member,
			MaxMember: list[i].MaxMember,
		}
	}
	return ret
}

func clubListHandler(form *nex.Form) (*protocol.ClubListResponse, error) {
	uid := form.Int64OrDefault("uid", -1)
	if uid < 0 {
		return nil, errors.New("服务器内部错误")
	}

	list := clubs(uid)
	return &protocol.ClubListResponse{Data: list}, nil
}

func thirdUserLoginHandler(r *http.Request, data *protocol.ThirdUserLoginRequest) (*protocol.LoginResponse, error) {
	logger.Infof("微信登录: %+v", data)
	if data == nil {
		return nil, errutil.ErrInvalidParameter
	}

	userInfo, err := oauth2.GetUserInfo(data.AccessToken, data.OpenID, "zh_CN", nil)
	if err != nil {
		return nil, err
	}

	logger.Debugf("微信用户信息: %+v", userInfo)

	var u *model.User
	thirdUser, err := db.QueryThirdAccount(data.OpenID, data.Platform)
	if err != nil && err != errutil.ErrThirdAccountNotFound {
		logger.Error(err)
		return nil, err
	}
	//用户存在
	if err == nil {
		u, err = db.QueryUser(thirdUser.Uid)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		// 更新昵称
		thirdUser.ThirdName = filterEmoji(userInfo.Nickname)
		thirdUser.HeadUrl = userInfo.HeadImageURL
		thirdUser.Sex = userInfo.Sex
		db.UpdateThirdAccount(thirdUser)
	} else {
		u = &model.User{Status: db.StatusNormal, IsOnline: db.UserOffline}
		u.Role = db.RoleTypeThird //角色类型
		u.Coin = defaultCoin

		thirdUser = &model.ThirdAccount{
			ThirdAccount: userInfo.OpenId,
			ThirdName:    filterEmoji(userInfo.Nickname),
			Platform:     data.Platform,
			HeadUrl:      userInfo.HeadImageURL,
			Sex:          userInfo.Sex,
		}

		if err := db.InsertThirdAccount(thirdUser, u); err != nil {
			return nil, err
		}

		db.RegisterUserLog(u, data.Device, data.AppID, data.ChannelID, protocol.RegTypeThird) //注册记录
	}

	checkSession(u.Id)

	resp := &protocol.LoginResponse{
		Name:     thirdUser.ThirdName,
		Uid:      u.Id, //注意此处是id而非uid
		HeadUrl:  thirdUser.HeadUrl,
		Sex:      thirdUser.Sex,
		IP:       host,
		Port:     port,
		FangKa:   u.Coin,
		PlayerIP: ip(r.RemoteAddr),
		Config:   config,
		Messages: messages,
		ClubList: clubs(u.Id),
		Debug:    0, //u.Debug,
	}

	// 插入登陆记录
	device := protocol.Device{
		IP:     ip(r.RemoteAddr),
		Remote: r.RemoteAddr,
	}
	db.InsertLoginLog(u.Id, device, data.AppID, data.ChannelID)

	return resp, nil
}

func guestLoginHandler(r *http.Request, data *protocol.LoginRequest) (*protocol.LoginResponse, error) {
	data.Device.IMEI = data.IMEI

	logger.Infof("游客登录IEMEI: %s", data.Device.IMEI)

	user, err := db.QueryGuestUser(data.AppID, data.Device.IMEI)
	if err != nil {
		// 生成一个新用户
		user = &model.User{
			Status:   db.StatusNormal,
			IsOnline: db.UserOffline,
			Role:     db.RoleTypeThird,
			Coin:     defaultCoin,
		}

		if err := db.InsertUser(user); err != nil {
			logger.Error(err.Error())
			return nil, err
		}

		db.RegisterUserLog(user, data.Device, data.AppID, data.ChannelID, protocol.RegTypeThird) //注册记录
	}

	checkSession(user.Id)

	resp := &protocol.LoginResponse{
		Uid:      user.Id,
		HeadUrl:  "http://wx.qlogo.cn/mmopen/s962LEwpLxhQSOnarDnceXjSxVGaibMRsvRM4EIWic0U6fQdkpqz4Vr8XS8D81QKfyYuwjwm2M2ibsFY8mia8ic51ww/0",
		Sex:      1,
		IP:       host,
		Port:     port,
		FangKa:   user.Coin,
		PlayerIP: ip(r.RemoteAddr),
		Config:   config,
		Messages: messages,
		ClubList: clubs(user.Id),
		Debug:    0, //user.Debug,
	}
	resp.Name = fmt.Sprintf("G%d", resp.Uid)

	// 插入登陆记录
	device := protocol.Device{
		IP:     ip(r.RemoteAddr),
		Remote: r.RemoteAddr,
	}
	db.InsertLoginLog(user.Id, device, data.AppID, data.ChannelID)

	return resp, nil
}

// 查询是否使用游客登陆
type (
	queryRequest struct {
		AppId     string `json:"appId"`
		ChannelId string `json:"channelId"`
	}

	queryResponse struct {
		Code  int  `json:"code"`
		Guest bool `json:"guest"`
	}
)

var (
	forbidGuest  = &queryResponse{Guest: false}
	accepetGuest = &queryResponse{Guest: true}
)

func queryHandler(query *queryRequest) (*queryResponse, error) {
	logger.Infof("%v", query)
	if !enableGuest {
		return forbidGuest, nil
	}

	for _, s := range guestChannels {
		if query.ChannelId == s {
			return accepetGuest, nil
		}
	}

	return forbidGuest, nil
}
