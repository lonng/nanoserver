package game

import (
	"time"

	"github.com/lonng/nano/scheduler"
	"github.com/lonng/nanoserver/pkg/constant"
	"github.com/lonng/nanoserver/protocol"
)

// 房间解散统计
type dissolveContext struct {
	desk     *Desk            // 牌桌
	status   map[int64]bool   //解散统计
	desc     map[int64]string //解散描述
	restTime int32            //解散剩余时间
	timer    *scheduler.Timer //取消解散房间
	pause    map[int64]bool   //离线状态
}

func newDissolveContext(desk *Desk) *dissolveContext {
	return &dissolveContext{
		desk:   desk,
		status: map[int64]bool{},
		desc:   map[int64]string{},
		pause:  map[int64]bool{},
	}
}

func (d *dissolveContext) reset() {
	d.status = map[int64]bool{}
	d.desc = map[int64]string{}
}

func (d *dissolveContext) stop() {
	if d.timer != nil {
		d.desk.logger.Info("关闭解散倒计时定时器")
		d.timer.Stop()
		d.timer = nil
	}
}

func (d *dissolveContext) start(restTime int32) {
	d.desk.logger.Debug("开始解散倒计时")

	//解散房间倒计时
	d.restTime = restTime
	d.timer = scheduler.NewTimer(time.Second, func() {
		if d.desk.status() == constant.DeskStatusDestory {
			d.desk.logger.Error("解散倒计时过程中已退出")
			d.stop()
			return
		}

		d.restTime--
		rest := d.restTime
		// 每30秒记录日志
		if rest%30 == 0 {
			d.desk.logger.Debugf("解散倒计时: %d", rest)
		}
		if rest < 0 {
			d.stop()
			d.desk.doDissolve()
			return
		}
	})
}

func (d *dissolveContext) isOnline(uid int64) bool {
	return !d.pause[uid]
}

func (d *dissolveContext) updateOnlineStatus(uid int64, online bool) {
	if online {
		delete(d.pause, uid)
	} else {
		d.pause[uid] = true
	}

	d.desk.logger.Debugf("玩家在线状态: %+v", d.pause)
	d.desk.group.Broadcast("onPlayerOfflineStatus", &protocol.PlayerOfflineStatus{Uid: uid, Offline: !online})
}

func (d *dissolveContext) setUidStatus(uid int64, agree bool, desc string) {
	d.status[uid] = agree
	d.desc[uid] = desc

	d.desk.logger.Debugf("玩家解散状态: %+v, %+v", d.status, d.desc)
}

// 是否已经是申请解散状态
func (d *dissolveContext) isDissolving() bool {
	return d.timer != nil
}

func (d *dissolveContext) agreeCount() int {
	return len(d.status)
}

func (d *dissolveContext) offlineCount() int {
	return len(d.pause)
}
