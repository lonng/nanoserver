//Package encoding encoding the error or response
package encoding

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"

	"github.com/lonnng/nanoserver/internal/errutil"
)

type errorer interface {
	error() error
}

func EncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		EncodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func EncodePlainResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		EncodeError(ctx, e.error(), w)
		return nil
	}

	var err error

	var data []byte
	switch response.(type) {
	case string:
		data = []byte(response.(string))

	case []byte:
		data = response.([]byte)

	default:
		return errutil.YXErrInvalidParameter
	}

	_, err = w.Write(data)
	return err
}

func EncodeError(ctx context.Context, e error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(encodeError(e))
}

func SimpleEncodeError(err error) interface{} {
	return encodeError(err)
}
