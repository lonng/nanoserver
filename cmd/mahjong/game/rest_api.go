package game

import (
	"github.com/lonng/nanoserver/protocol"
)

// TODO: conc
func Kick(uid int64) error {
	defaultManager.chKick <- uid
	return nil
}

func BroadcastSystemMessage(message string) {
	defaultManager.group.Broadcast("onBroadcast", &protocol.StringMessage{Message: message})
}

func Reset(uid int64) {
	defaultManager.chReset <- uid
}

func Recharge(uid, coin int64) {
	defaultManager.chRecharge <- RechargeInfo{uid, coin}
}
