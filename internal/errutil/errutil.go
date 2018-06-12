package errutil

import (
	"errors"
)

var (
	YXErrBadRoute              = errors.New("bad route")
	YXErrWrongType             = errors.New("wrong type")
	YXErrNotFound              = errors.New("not found")
	YXErrUserNameExists        = errors.New("user name exists")
	YXErrIllegalParameter      = errors.New("illegal parameter")
	YXErrDBOperation           = errors.New("database opertaion failed")
	YXErrNotRSAPublicKey       = errors.New("not a rsa public key")
	YXErrNotRSAPrivateKey      = errors.New("not a rsa private key")
	YXErrAuthFailed            = errors.New("auth failed")
	YXErrIllegalLoginType      = errors.New("illegal login type")
	YXErrWrongPassword         = errors.New("wrong password")
	YXErrUserNameNotFound      = errors.New("username not found")
	YXErrInitFailed            = errors.New("initialize failed")
	YXErrServerInternal        = errors.New("server internal error")
	YXErrCacheOperation        = errors.New("cache opertaion failed")
	YXErrPermissionDenied      = errors.New("permission denied")
	YXErrNotImplemented        = errors.New("not implemented")
	YXErrUserNotFound          = errors.New("user not found")
	YXErrTokenNotFound         = errors.New("token not found")
	YXErrInvalidToken          = errors.New("invalid token")
	YXErrIllegalName           = errors.New("illegal user name")
	YXErrTokenMismatchUser     = errors.New("token mismatch user")
	YXErrInvalidPayPlatform    = errors.New("invalid pay platform")
	YXErrUnsupportSignType     = errors.New("unsupport sign type")
	YXErrOrderNotFound         = errors.New("order not found")
	YXErrInvalidParameter      = errors.New("invalid parameter")
	YXErrRequestFailed         = errors.New("request failed")
	YXErrIllegalOrderType      = errors.New("illegal order type")
	YXErrCoinNotEnough         = errors.New("youxian coin not enough")
	YXErrFrequencyLimited      = errors.New("frequency limited")
	YXErrVerifyFailed          = errors.New("verify failed")
	YXErrSignFailed            = errors.New("sign failed")
	YXErrTradeExisted          = errors.New("trade has existed")
	YXErrProviderNotFound      = errors.New("provider not found")
	YXErrThirdAccountNotFound  = errors.New("third account not found")
	YXErrWrongThirdLoginType   = errors.New("wrong third login type")
	YXErrDirNotExists          = errors.New("directory not exists")
	YXErrPayTestDisable        = errors.New("pay test disable")
	YXErrPropertyNotFound      = errors.New("property not found")
	YXErrIllegalDeskStatus     = errors.New("illegal desk status")
	YXErrPlayerNotFound        = errors.New("player not found")
	YXErrNoSuchWinPoints       = errors.New("no such win points")
	YXErrDismatchTileNum       = errors.New("a shortage or surplus of tiles")
	YXErrNotWon                = errors.New("not won now")
	YXErrDeskNotFound          = errors.New("desk not found")
	YXErrUUIDNotFound          = errors.New("uuid not found")
	YXErrProductionNotFound    = errors.New("production not found")
	YXErrRequestPrePayIDFailed = errors.New("request prepay id failed")
	ErrAccountExists           = errors.New("account exists")
)

//Code code for the error
func Code(err error) int {
	if c, ok := errs[err]; ok {
		return c
	}
	return Unknown
}
