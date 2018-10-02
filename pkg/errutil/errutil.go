package errutil

import (
	"errors"
)

var (
	ErrBadRoute              = errors.New("bad route")
	ErrWrongType             = errors.New("wrong type")
	ErrNotFound              = errors.New("not found")
	ErrUserNameExists        = errors.New("user name exists")
	ErrIllegalParameter      = errors.New("illegal parameter")
	ErrDBOperation           = errors.New("database opertaion failed")
	ErrNotRSAPublicKey       = errors.New("not a rsa public key")
	ErrNotRSAPrivateKey      = errors.New("not a rsa private key")
	ErrAuthFailed            = errors.New("auth failed")
	ErrIllegalLoginType      = errors.New("illegal login type")
	ErrWrongPassword         = errors.New("wrong password")
	ErrUserNameNotFound      = errors.New("username not found")
	ErrInitFailed            = errors.New("initialize failed")
	ErrServerInternal        = errors.New("server internal error")
	ErrCacheOperation        = errors.New("cache opertaion failed")
	ErrPermissionDenied      = errors.New("permission denied")
	ErrNotImplemented        = errors.New("not implemented")
	ErrUserNotFound          = errors.New("user not found")
	ErrTokenNotFound         = errors.New("token not found")
	ErrInvalidToken          = errors.New("invalid token")
	ErrIllegalName           = errors.New("illegal user name")
	ErrTokenMismatchUser     = errors.New("token mismatch user")
	ErrInvalidPayPlatform    = errors.New("invalid pay platform")
	ErrUnsupportSignType     = errors.New("unsupport sign type")
	ErrOrderNotFound         = errors.New("order not found")
	ErrInvalidParameter      = errors.New("invalid parameter")
	ErrRequestFailed         = errors.New("request failed")
	ErrIllegalOrderType      = errors.New("illegal order type")
	ErrCoinNotEnough         = errors.New("youxian coin not enough")
	ErrFrequencyLimited      = errors.New("frequency limited")
	ErrVerifyFailed          = errors.New("verify failed")
	ErrSignFailed            = errors.New("sign failed")
	ErrTradeExisted          = errors.New("trade has existed")
	ErrProviderNotFound      = errors.New("provider not found")
	ErrThirdAccountNotFound  = errors.New("third account not found")
	ErrWrongThirdLoginType   = errors.New("wrong third login type")
	ErrDirNotExists          = errors.New("directory not exists")
	ErrPayTestDisable        = errors.New("pay test disable")
	ErrPropertyNotFound      = errors.New("property not found")
	ErrIllegalDeskStatus     = errors.New("illegal desk status")
	ErrPlayerNotFound        = errors.New("player not found")
	ErrNoSuchWinPoints       = errors.New("no such win points")
	ErrDismatchTileNum       = errors.New("a shortage or surplus of tiles")
	ErrNotWon                = errors.New("not won now")
	ErrDeskNotFound          = errors.New("desk not found")
	ErrUUIDNotFound          = errors.New("uuid not found")
	ErrProductionNotFound    = errors.New("production not found")
	ErrRequestPrePayIDFailed = errors.New("request prepay id failed")
	ErrAccountExists         = errors.New("account exists")
)

//Code code for the error
func Code(err error) int {
	if c, ok := errs[err]; ok {
		return c
	}
	return Unknown
}
