package errutil

const (
	codeBase = 1000
)

const (
	Unknown = codeBase + iota
	yxBadRoute
	yxNotFound
	yxWrongType
	yxUserNameExists
	yxAuthFailed
	yxIllegalParameter
	yxNotRSAPublicKey
	yxNotRSAPrivateKey
	yxIllegalLoginType
	yxWrongPassword
	yxUserNameNotFound
	yxInitFailed
	yxServerInternal
	yxDbOperation
	yxCacheOperation
	yxPermissionDenied
	yxNotImplemented

	yxUserNotFound
	yxTokenNotFound
	yxInvalidToken
	yxIllegalName
	yxTokenMismatchUser
	yxInvalidPlatform
	yxUnsupportSignType
	yxOrderNotFound
	yxInvalidParameter
	yxRequestFailed
	yxIllegalOrderType
	yxCoinNotEnough
	yxFrequencyLimited
	yxVerifyFailed
	yxSignFailed
	yxTradeExisted
	yxProviderNotFound
	yxThirdAccountNotFound
	yxWrongThirdLoginType
	yxDirNotExists
	yxPayTestDisable
	yxPropertyNotFound
	yxUuidNotFound
	yxProductionNotFound
	yxRequestPrePayIDFailed
	YXDeskNotFound
)

var errs = map[error]int{
	ErrBadRoute:              yxBadRoute,
	ErrNotFound:              yxNotFound,
	ErrWrongType:             yxWrongType,
	ErrUserNameExists:        yxUserNameExists,
	ErrIllegalParameter:      yxIllegalParameter,
	ErrNotRSAPublicKey:       yxNotRSAPublicKey,
	ErrNotRSAPrivateKey:      yxNotRSAPrivateKey,
	ErrAuthFailed:            yxAuthFailed,
	ErrIllegalLoginType:      yxIllegalLoginType,
	ErrWrongPassword:         yxWrongPassword,
	ErrUserNameNotFound:      yxUserNameNotFound,
	ErrInitFailed:            yxInitFailed,
	ErrServerInternal:        yxServerInternal,
	ErrDBOperation:           yxDbOperation,
	ErrCacheOperation:        yxCacheOperation,
	ErrPermissionDenied:      yxPermissionDenied,
	ErrNotImplemented:        yxNotImplemented,
	ErrUserNotFound:          yxUserNotFound,
	ErrTokenNotFound:         yxTokenNotFound,
	ErrInvalidToken:          yxInvalidToken,
	ErrIllegalName:           yxIllegalName,
	ErrTokenMismatchUser:     yxTokenMismatchUser,
	ErrInvalidPayPlatform:    yxInvalidPlatform,
	ErrUnsupportSignType:     yxUnsupportSignType,
	ErrOrderNotFound:         yxOrderNotFound,
	ErrInvalidParameter:      yxInvalidParameter,
	ErrRequestFailed:         yxRequestFailed,
	ErrIllegalOrderType:      yxIllegalOrderType,
	ErrCoinNotEnough:         yxCoinNotEnough,
	ErrFrequencyLimited:      yxFrequencyLimited,
	ErrVerifyFailed:          yxVerifyFailed,
	ErrSignFailed:            yxSignFailed,
	ErrTradeExisted:          yxTradeExisted,
	ErrProviderNotFound:      yxProviderNotFound,
	ErrThirdAccountNotFound:  yxThirdAccountNotFound,
	ErrWrongThirdLoginType:   yxWrongThirdLoginType,
	ErrDirNotExists:          yxDirNotExists,
	ErrPayTestDisable:        yxPayTestDisable,
	ErrPropertyNotFound:      yxPropertyNotFound,
	ErrUUIDNotFound:          yxUuidNotFound,
	ErrProductionNotFound:    yxProductionNotFound,
	ErrRequestPrePayIDFailed: yxRequestPrePayIDFailed,
	ErrDeskNotFound:          YXDeskNotFound,
}
