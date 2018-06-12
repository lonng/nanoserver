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
	YXErrBadRoute:              yxBadRoute,
	YXErrNotFound:              yxNotFound,
	YXErrWrongType:             yxWrongType,
	YXErrUserNameExists:        yxUserNameExists,
	YXErrIllegalParameter:      yxIllegalParameter,
	YXErrNotRSAPublicKey:       yxNotRSAPublicKey,
	YXErrNotRSAPrivateKey:      yxNotRSAPrivateKey,
	YXErrAuthFailed:            yxAuthFailed,
	YXErrIllegalLoginType:      yxIllegalLoginType,
	YXErrWrongPassword:         yxWrongPassword,
	YXErrUserNameNotFound:      yxUserNameNotFound,
	YXErrInitFailed:            yxInitFailed,
	YXErrServerInternal:        yxServerInternal,
	YXErrDBOperation:           yxDbOperation,
	YXErrCacheOperation:        yxCacheOperation,
	YXErrPermissionDenied:      yxPermissionDenied,
	YXErrNotImplemented:        yxNotImplemented,
	YXErrUserNotFound:          yxUserNotFound,
	YXErrTokenNotFound:         yxTokenNotFound,
	YXErrInvalidToken:          yxInvalidToken,
	YXErrIllegalName:           yxIllegalName,
	YXErrTokenMismatchUser:     yxTokenMismatchUser,
	YXErrInvalidPayPlatform:    yxInvalidPlatform,
	YXErrUnsupportSignType:     yxUnsupportSignType,
	YXErrOrderNotFound:         yxOrderNotFound,
	YXErrInvalidParameter:      yxInvalidParameter,
	YXErrRequestFailed:         yxRequestFailed,
	YXErrIllegalOrderType:      yxIllegalOrderType,
	YXErrCoinNotEnough:         yxCoinNotEnough,
	YXErrFrequencyLimited:      yxFrequencyLimited,
	YXErrVerifyFailed:          yxVerifyFailed,
	YXErrSignFailed:            yxSignFailed,
	YXErrTradeExisted:          yxTradeExisted,
	YXErrProviderNotFound:      yxProviderNotFound,
	YXErrThirdAccountNotFound:  yxThirdAccountNotFound,
	YXErrWrongThirdLoginType:   yxWrongThirdLoginType,
	YXErrDirNotExists:          yxDirNotExists,
	YXErrPayTestDisable:        yxPayTestDisable,
	YXErrPropertyNotFound:      yxPropertyNotFound,
	YXErrUUIDNotFound:          yxUuidNotFound,
	YXErrProductionNotFound:    yxProductionNotFound,
	YXErrRequestPrePayIDFailed: yxRequestPrePayIDFailed,
	YXErrDeskNotFound: YXDeskNotFound,

}
