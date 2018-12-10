package algoutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/lonng/nanoserver/pkg/errutil"
)

const (
	argSep = "&"
	kvSep  = "="
)

func ParseParams(params string) map[string]string {
	m := map[string]string{}
	parts := strings.Split(params, argSep)
	for _, arg := range parts {
		// strings.Split(s, "=") will cause error when signature has
		// padding(that is something like "==")
		i := strings.IndexAny(arg, kvSep)
		if i < 0 {
			continue
		}
		k := arg[:i]
		v := arg[i+1:]
		m[k] = v
	}
	return m
}

// TODO: WARNING!!!!!!!
// All struct field must be string type
func ParamsToStruct(params string, v interface{}) error {
	if params == "" || len(strings.TrimSpace(params)) < 1 {
		return errutil.ErrIllegalParameter
	}
	m := ParseParams(params)
	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, v)
}

func SortParams(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	var (
		i    = 0
		keys = make([]string, len(m))
		buf  = &bytes.Buffer{}
	)
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(m[k])
		buf.WriteString("&")
	}
	buf.Truncate(buf.Len() - 1)
	return buf.String()
}

func ConcatWithURLEncode(params map[string]string) *bytes.Buffer {
	if params == nil || len(params) == 0 {
		return nil
	}

	buf := &bytes.Buffer{}
	for k, v := range params {
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(url.QueryEscape(v))
		buf.WriteString("&")
	}

	buf.Truncate(buf.Len() - 1)

	return buf
}

// SortAndConcat sort the map by key in ASCII order,
// and concat it in form of "k1=v1&k2=v2"
func SortAndConcat(params map[string]string, extras ...bool) []byte {
	if params == nil || len(params) == 0 {
		return nil
	}

	var trimSpace bool = true
	if len(extras) > 0 {
		trimSpace = extras[0]
	}

	keys := make([]string, len(params))
	i := 0
	for k := range params {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	buf := &bytes.Buffer{}
	for _, k := range keys {
		if trimSpace && params[k] == "" {
			continue
		}
		buf.WriteString(fmt.Sprintf("%s=%s&", k, params[k]))

	}
	buf.Truncate(buf.Len() - 1)
	return buf.Bytes()
}
