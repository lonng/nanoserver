package nex

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

type testRequest struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}
type testResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var successResponse = &testResponse{Message: "success"}

// acceptable function signature
func withNone() (*testResponse, error)                         { return successResponse, nil }
func withBody(io.ReadCloser) (*testResponse, error)            { return successResponse, nil }
func withReq(*testRequest) (*testResponse, error)              { return successResponse, nil }
func withHeader(http.Header) (*testResponse, error)            { return successResponse, nil }
func withForm(Form) (*testResponse, error)                     { return successResponse, nil }
func withPostForm(PostForm) (*testResponse, error)             { return successResponse, nil }
func withFormPtr(*Form) (*testResponse, error)                 { return successResponse, nil }
func withPostFormPtr(*PostForm) (*testResponse, error)         { return successResponse, nil }
func withMultipartForm(*multipart.Form) (*testResponse, error) { return successResponse, nil }
func withUrl(*url.URL) (*testResponse, error)                  { return successResponse, nil }
func withRawRequest(*http.Request) (*testResponse, error)      { return successResponse, nil }

func withMulti(*testRequest, Form, PostForm, http.Header, *url.URL) (*testResponse, error) {
	return nil, nil
}
func withAll(io.ReadCloser, *testRequest, Form, PostForm, http.Header, *multipart.Form, *url.URL) (*testResponse, error) {
	return nil, nil
}

func TestHandler(t *testing.T) {
	Handler(withNone)
	Handler(withBody)
	Handler(withReq)
	Handler(withHeader)
	Handler(withForm)
	Handler(withPostForm)
	Handler(withFormPtr)
	Handler(withPostFormPtr)
	Handler(withMultipartForm)
	Handler(withUrl)
	Handler(withRawRequest)
	Handler(withMulti)
	Handler(withAll)
}

func BenchmarkSimplePlainAdapter_Invoke(b *testing.B) {
	adapter := &simplePlainAdapter{reflect.ValueOf(withNone)}
	request, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		adapter.Invoke(recorder, request)
	}
}

func BenchmarkSimpleUnaryAdapter_Invoke(b *testing.B) {
	adapter := &simpleUnaryAdapter{reflect.TypeOf(&testRequest{}), reflect.ValueOf(withReq)}
	request, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		b.Fatal(err)
	}
	payload := []byte(`{"for":"hello", "bar":10000}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
		recorder := httptest.NewRecorder()
		adapter.Invoke(recorder, request)
	}
}

func BenchmarkGenericAdapter_Invoke(b *testing.B) {
	adapter := makeGenericAdapter(reflect.ValueOf(withMulti))
	request, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		b.Fatal(err)
	}
	payload := []byte(`{"for":"hello", "bar":10000}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
		recorder := httptest.NewRecorder()
		adapter.Invoke(recorder, request)
	}
}

func BenchmarkSimplePlainAdapter_Invoke2(b *testing.B) {
	handler := &Nex{adapter: &simplePlainAdapter{reflect.ValueOf(withNone)}}
	request, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
	}
}

func BenchmarkSimpleUnaryAdapter_Invoke2(b *testing.B) {
	handler := &Nex{adapter: &simpleUnaryAdapter{reflect.TypeOf(&testRequest{}), reflect.ValueOf(withReq)}}
	request, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		b.Fatal(err)
	}
	payload := []byte(`{"for":"hello", "bar":10000}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
	}
}

func BenchmarkGenericAdapter_Invoke2(b *testing.B) {
	handler := &Nex{adapter: makeGenericAdapter(reflect.ValueOf(withMulti))}
	request, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		b.Fatal(err)
	}
	payload := []byte(`{"for":"hello", "bar":10000}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
	}
}
