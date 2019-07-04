package game

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/lonng/nano"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/pipeline"
	"github.com/lonng/nano/serialize/json"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	version     = ""            // 游戏版本
	consume     = map[int]int{} // 房卡消耗配置
	forceUpdate = false
	logger      = log.WithField("component", "game")
)

// SetCardConsume 设置房卡消耗数量
func SetCardConsume(cfg string) {
	for _, c := range strings.Split(cfg, ",") {
		parts := strings.Split(c, "/")
		if len(parts) < 2 {
			logger.Warnf("无效的房卡配置: %s", c)
			continue
		}
		round, card := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		rd, err := strconv.Atoi(round)
		if err != nil {
			continue
		}
		cd, err := strconv.Atoi(card)
		if err != nil {
			continue
		}
		consume[rd] = cd
	}

	logger.Infof("当前游戏房卡消耗配置: %+v", consume)
}

// Startup 初始化游戏服务器
func Startup() {
	rand.Seed(time.Now().Unix())
	version = viper.GetString("update.version")

	heartbeat := viper.GetInt("core.heartbeat")
	if heartbeat < 5 {
		heartbeat = 5
	}

	// 房卡消耗配置
	csm := viper.GetString("core.consume")
	SetCardConsume(csm)
	forceUpdate = viper.GetBool("update.force")

	logger.Infof("当前游戏服务器版本: %s, 是否强制更新: %t, 当前心跳时间间隔: %d秒", version, forceUpdate, heartbeat)
	logger.Info("game service starup")

	// register game handler
	comps := &component.Components{}
	comps.Register(defaultManager)
	comps.Register(defaultDeskManager)
	comps.Register(new(ClubManager))

	// 加密管道
	c := newCrypto()
	pip := pipeline.New()
	pip.Inbound().PushBack(c.inbound)
	pip.Outbound().PushBack(c.outbound)

	addr := fmt.Sprintf(":%d", viper.GetInt("game-server.port"))
	nano.Listen(addr,
		nano.WithPipeline(pip),
		nano.WithHeartbeatInterval(time.Duration(heartbeat)*time.Second),
		nano.WithLogger(log.WithField("component", "nano")),
		nano.WithSerializer(json.NewSerializer()),
		nano.WithComponents(comps),
	)
}
