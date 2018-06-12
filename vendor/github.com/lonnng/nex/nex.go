package nex

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
)

type ErrorEncoder func(error) interface{}

type DefaultErrorMessage struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type Nex struct {
	before  []BeforeFunc
	after   []AfterFunc
	adapter HandlerAdapter
}

var errorEncoder ErrorEncoder

func fail(w http.ResponseWriter, err error) {
	errMsg := errorEncoder(err)
	//log.Infof("Result=Failed, Error=%v", errMsg)
	json.NewEncoder(w).Encode(errMsg)
}

func succ(w http.ResponseWriter, data interface{}) {
	//log.Infof("Result=Success, Response=%+v", data)
	json.NewEncoder(w).Encode(data)
}

func (n *Nex) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var (
		ctx  context.Context = context.Background()
		err  error
		resp interface{}
	)
	// global before middleware
	for _, b := range globalBefore {
		ctx, err = b(ctx, r)
		if err != nil {
			fail(w, err)
			return
		}
	}

	// before middleware
	for _, b := range n.before {
		ctx, err = b(ctx, r)
		if err != nil {
			fail(w, err)
			return
		}
	}

	// adapter handler
	resp, err = n.adapter.Invoke(w, r)

	// after middleware
	for _, a := range n.after {
		ctx, err = a(ctx, w)
		if err != nil {
			fail(w, err)
			return
		}
	}

	// global after middleware
	for _, a := range globalAfter {
		ctx, err = a(ctx, w)
		if err != nil {
			fail(w, err)
			return
		}
	}
	if err != nil {
		fail(w, err)
	} else {
		succ(w, resp)
	}
}

func (n *Nex) Before(before ...BeforeFunc) *Nex {
	for _, b := range before {
		if b != nil {
			n.before = append(n.before, b)
		}
	}
	return n
}

func (n *Nex) After(after ...AfterFunc) *Nex {
	for _, a := range after {
		if a != nil {
			n.after = append(n.after, a)
		}
	}
	return n
}

func Handler(f interface{}) *Nex {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		panic("invalid parameter")
	}

	if t.NumOut() != 2 {
		panic("unsupport function type, function return values should contain response data & error")
	}

	var adapter HandlerAdapter
	var num = t.NumIn()

	if num == 0 {
		adapter = &simplePlainAdapter{reflect.ValueOf(f)}
	} else if num == 1 && !isSupportType(t.In(0)) && t.In(0).Kind() == reflect.Ptr {
		adapter = &simpleUnaryAdapter{t.In(0), reflect.ValueOf(f)}
	} else {
		adapter = makeGenericAdapter(reflect.ValueOf(f))
	}

	return &Nex{adapter: adapter}
}

func SetErrorEncoder(c ErrorEncoder) {
	if c == nil {
		panic("nil pointer to error encoder")
	}
	errorEncoder = c
}

func SetMultipartFormMaxMemory(m int64) {
	maxMemory = m
}

func init() {
	errorEncoder = func(err error) interface{} {
		return &DefaultErrorMessage{
			Code:  -1,
			Error: err.Error(),
		}
	}
}
