package nex

import (
	"context"
	"net/http"
	"testing"
)

func before1(ctx context.Context, r *http.Request)(context.Context, error) {
	return context.WithValue(ctx, "testkey", "testvalue"), nil
}

// TODO
func TestNex_Before(t *testing.T) {

}
