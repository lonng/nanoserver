// +build debug

//Package encoding encoding the error or response
package encoding

import (
	kithttp "github.com/go-kit/kit/transport/http"
)

func encodeError(e error) interface{} {
	var response protocol.ErrorResponse
	if raw, ok := e.(kithttp.Error); ok {
		response = protocol.ErrorResponse{Code: errutil.Code(raw.Err), Error: raw.Err.Error()}
	} else {
		response = protocol.ErrorResponse{Code: errutil.Code(e), Error: e.Error()}
	}
	return response
}
