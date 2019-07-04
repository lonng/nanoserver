package web

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/lonng/nanoserver/db"
	"github.com/lonng/nanoserver/internal/game"
	"github.com/lonng/nanoserver/internal/web/api"
	"github.com/lonng/nanoserver/pkg/errutil"
	"github.com/lonng/nanoserver/protocol"
	"github.com/lonng/nex"
	log "github.com/sirupsen/logrus"
)

func authFilter(_ context.Context, r *http.Request) (context.Context, error) {
	parts := strings.Split(r.RemoteAddr, ":")
	if len(parts) < 2 {
		return context.Background(), errutil.ErrPermissionDenied
	}

	if parts[0] != "127.0.0.1" {
		return context.Background(), errutil.ErrPermissionDenied
	}

	return context.Background(), nil
}

func broadcast(query *nex.Form) (*protocol.StringMessage, error) {
	message := strings.TrimSpace(query.Get("message"))
	if message == "" || len(message) < 5 {
		return nil, errors.New("消息不可小于5个字")
	}
	api.AddMessage(message)
	game.BroadcastSystemMessage(message)
	return protocol.SuccessMessage, nil
}

func resetPlayerHandler(query *nex.Form) (*protocol.StringMessage, error) {
	uid := query.Int64OrDefault("uid", -1)
	if uid <= 0 {
		return nil, errutil.ErrIllegalParameter
	}
	log.Infof("手动重置玩家数据: Uid=%d", uid)
	game.Reset(uid)
	return protocol.SuccessMessage, nil
}

func kickHandler(query *nex.Form) (*protocol.StringMessage, error) {
	uid := query.Int64OrDefault("uid", -1)
	if uid <= 0 {
		return nil, errutil.ErrIllegalParameter
	}

	log.Infof("踢玩家下线: Uid=%d", uid)
	if err := game.Kick(uid); err != nil {
		return nil, err
	}

	return protocol.SuccessMessage, nil
}

func onlineHandler(query *nex.Form) (interface{}, error) {
	begin := query.Int64OrDefault("begin", 0)
	end := query.Int64OrDefault("end", -1)
	if end < 0 {
		end = time.Now().Unix()
	}

	log.Infof("获取在线数据信息: begin=%s, end=%s", time.Unix(begin, 0).String(), time.Unix(end, 0).String())
	return db.OnlineStats(begin, end)
}

func rechargeHandler(data *protocol.RechargeRequest) (*protocol.StringMessage, error) {
	if data.Uid < 1 || data.Count < 1 {
		return nil, errutil.ErrIllegalParameter
	}
	u, err := db.QueryUser(data.Uid)
	if err != nil {
		return nil, err
	}

	u.Coin += data.Count

	if err := db.UpdateUser(u); err != nil {
		return nil, err
	}

	// 通知客户端
	game.Recharge(u.Id, u.Coin)

	log.Infof("给玩家充值: Uid=%d, end=%d", data.Uid, data.Count)
	return protocol.SuccessMessage, nil
}

// http://127.0.0.1:12306/v1/gm/consume?consume="4/1,8/1,16/2"
func cardConsumeHandler(query *nex.Form) (*protocol.StringMessage, error) {
	consume := query.Get("consume")
	if consume == "" {
		return nil, errutil.ErrIllegalParameter
	}
	log.Infof("手动重置房卡消耗数据: %s", consume)
	game.SetCardConsume(consume)
	return protocol.SuccessMessage, nil
}
func userInfoHandler(query *nex.Form) (interface{}, error) {
	id := query.Int64OrDefault("id", -1)
	if id <= 0 {
		return nil, errutil.ErrIllegalParameter
	}

	return db.QueryUserInfo(id)
}
