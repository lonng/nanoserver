// +build debug

package http

import (
	"fmt"
	"net/http"
)

func (s Server) detail(r *http.Request, request, response interface{}) {
	s.logger.Log(
		"method", r.Method,
		"url", r.URL.String(),
		"request", fmt.Sprintf("%+v", request),
		"response", fmt.Sprintf("%+v", response))
}
