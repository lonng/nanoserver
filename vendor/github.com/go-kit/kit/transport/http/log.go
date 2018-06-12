// +build !debug

package http

import (
	"net/http"
)

func (s Server) detail(r *http.Request, request, response interface{}) {}
