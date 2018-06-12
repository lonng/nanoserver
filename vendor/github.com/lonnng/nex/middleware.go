package nex

import (
	"context"
	"net/http"
)

// Global middleware, such as IP filter, logs
var (
	globalBefore []BeforeFunc
	globalAfter  []AfterFunc
)

type BeforeFunc func(context.Context, *http.Request) (context.Context, error)
type AfterFunc func(context.Context, http.ResponseWriter) (context.Context, error)

func Before(before ...BeforeFunc) {
	for _, b := range before {
		if b != nil {
			globalBefore = append(globalBefore, b)
		}
	}
}

func After(after ...AfterFunc) {
	for _, a := range after {
		if a != nil {
			globalAfter = append(globalAfter, a)
		}
	}
}
