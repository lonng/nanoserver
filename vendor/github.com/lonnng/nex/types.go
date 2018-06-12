package nex

import (
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
)

type valuer func(r *http.Request) reflect.Value

var supportTypes = map[reflect.Type]valuer{
	reflect.TypeOf((*io.ReadCloser)(nil)).Elem(): bodyValuer,        // request.Body
	reflect.TypeOf((http.Header)(nil)):           headerValuer,      // request.Header
	reflect.TypeOf(Form{}):                       formValuer,        // request.Form
	reflect.TypeOf(PostForm{}):                   postFromValuer,    // request.PostFrom
	reflect.TypeOf((*Form)(nil)):                 formPtrValuer,     // request.Form
	reflect.TypeOf((*PostForm)(nil)):             postFromPtrValuer, // request.PostFrom
	reflect.TypeOf((*url.URL)(nil)):              urlValuer,         // request.URL
	reflect.TypeOf((*multipart.Form)(nil)):       multipartValuer,   // request.MultipartForm
	reflect.TypeOf((*http.Request)(nil)):         requestValuer,     // raw request
}

var maxMemory = int64(2 * 1024 * 1024)

type uniform struct {
	url.Values
}

type Form struct {
	uniform
}

type PostForm struct {
	uniform
}

func bodyValuer(r *http.Request) reflect.Value {
	return reflect.ValueOf(r.Body)
}

func urlValuer(r *http.Request) reflect.Value {
	return reflect.ValueOf(r.URL)
}

func headerValuer(r *http.Request) reflect.Value {
	return reflect.ValueOf(r.Header)
}

func multipartValuer(r *http.Request) reflect.Value {
	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		panic(err)
	}
	return reflect.ValueOf(r.MultipartForm)
}

func formValuer(r *http.Request) reflect.Value {
	r.ParseForm()
	return reflect.ValueOf(Form{uniform{r.Form}})
}

func postFromValuer(r *http.Request) reflect.Value {
	r.ParseForm()
	return reflect.ValueOf(PostForm{uniform{r.PostForm}})
}

func formPtrValuer(r *http.Request) reflect.Value {
	r.ParseForm()
	return reflect.ValueOf(&Form{uniform{r.Form}})
}

func postFromPtrValuer(r *http.Request) reflect.Value {
	r.ParseForm()
	return reflect.ValueOf(&PostForm{uniform{r.PostForm}})
}

func requestValuer(r *http.Request) reflect.Value {
	return reflect.ValueOf(r)
}

func isSupportType(t reflect.Type) bool {
	_, ok := supportTypes[t]
	return ok
}
