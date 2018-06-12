package nex

import (
	"encoding/json"
	"net/http"
	"reflect"
)

type HandlerAdapter interface {
	Invoke(http.ResponseWriter, *http.Request) (interface{}, error)
}

type genericAdapter struct {
	method reflect.Value
	numIn  int
	types  []reflect.Type
}

// Accept zero parameter adapter
type simplePlainAdapter struct {
	method reflect.Value
}

// Accept only one parameter adapter
type simpleUnaryAdapter struct {
	argType reflect.Type
	method  reflect.Value
}

func makeGenericAdapter(method reflect.Value) *genericAdapter {
	var noSupportExists = false
	t := method.Type()
	numIn := t.NumIn()

	a := &genericAdapter{
		method: method,
		numIn:  numIn,
		types:  make([]reflect.Type, numIn),
	}

	for i := 0; i < numIn; i++ {
		in := t.In(i)
		if !isSupportType(in) {
			if noSupportExists {
				panic("function should accept only one customize type")
			}

			if in.Kind() != reflect.Ptr {
				panic("customize type should be a pointer(" + in.PkgPath() + "." + in.Name() + ")")
			}
			noSupportExists = true
		}
		a.types[i] = in
	}

	return a
}

func (a *genericAdapter) Invoke(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	values := make([]reflect.Value, a.numIn)
	for i := 0; i < a.numIn; i++ {
		v, ok := supportTypes[a.types[i]]
		if ok {
			values[i] = v(r)
		} else {
			d := reflect.New(a.types[i].Elem()).Interface()
			err := json.NewDecoder(r.Body).Decode(d)
			if err != nil {
				return nil, err
			}
			values[i] = reflect.ValueOf(d)
		}
	}

	ret := a.method.Call(values)
	if err := ret[1].Interface(); err != nil {
		return nil, err.(error)
	}

	return ret[0].Interface(), nil
}

func (a *simplePlainAdapter) Invoke(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ret := a.method.Call([]reflect.Value{})
	if err := ret[1].Interface(); err != nil {
		return nil, err.(error)
	}

	return ret[0].Interface(), nil
}

func (a *simpleUnaryAdapter) Invoke(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	data := reflect.New(a.argType.Elem()).Interface()
	err := json.NewDecoder(r.Body).Decode(data)
	if err != nil {
		return nil, err
	}

	ret := a.method.Call([]reflect.Value{reflect.ValueOf(data)})
	if err := ret[1].Interface(); err != nil {
		return nil, err.(error)
	}

	return ret[0].Interface(), nil
}
