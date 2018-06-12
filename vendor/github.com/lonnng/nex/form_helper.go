package nex

import (
	"strconv"
)

func (f *uniform) Int(key string) int {
	value := f.Get(key)
	if value == "" {
		return 0
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return v
}

func (f *uniform) IntOrDefault(key string, def int) int {
	value := f.Get(key)
	if value == "" {
		return def
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return def
	}
	return v
}

func (f *uniform) Int64(key string) int64 {
	value := f.Get(key)
	if value == "" {
		return 0
	}
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return v
}

func (f *uniform) Int64OrDefault(key string, def int64) int64 {
	value := f.Get(key)
	if value == "" {
		return def
	}
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return def
	}
	return v
}

func (f *uniform) Uint64(key string) uint64 {
	value := f.Get(key)
	if value == "" {
		return 0
	}
	v, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}
	return v
}

func (f *uniform) Uint64OrDefault(key string, def uint64) uint64 {
	value := f.Get(key)
	if value == "" {
		return def
	}
	v, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return def
	}
	return v
}

// Get gets the first value associated with the given key.
// If there are no values associated with the key, Get returns
// the empty string. To access multiple values, use the map
// directly.
func (f *uniform) Get(key string) string {
	return f.Values.Get(key)
}

// Set sets the key to value. It replaces any existing
// values.
func (f *uniform) Set(key, value string) {
	f.Values.Set(key,value)
}

// Add adds the value to key. It appends to any existing
// values associated with key.
func (f *uniform) Add(key, value string) {
	f.Values.Add(key,value)
}

// Del deletes the values associated with key.
func (f *uniform) Del(key string) {
	f.Values.Del(key)
}


// Encode encodes the values into ``URL encoded'' form
// ("bar=baz&foo=quux") sorted by key.
func (f *uniform) Encode() string {
	return f.Values.Encode()
}