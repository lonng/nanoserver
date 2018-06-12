// +build !debug

package encoding

import (
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/lonnng/nanoserver/internal/errutil"
	"github.com/lonnng/nanoserver/internal/protocol"
)

func encodeError(e error) interface{} {
	var response protocol.ErrorResponse
	var (
		code = errutil.Code(e)
		err  = e.Error()
	)
	if raw, ok := e.(kithttp.Error); ok {
		code = errutil.Code(raw.Err)
		err = raw.Err.Error()
	}

	if code == errutil.Unknown {
		err = errutil.YXErrServerInternal.Error()
	}
	response = protocol.ErrorResponse{Code: code, Error: err}

	return response
}
