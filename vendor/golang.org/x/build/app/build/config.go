// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build appengine

package build

import (
	"sync"

	"appengine"
	"appengine/datastore"
)

// A global map of rarely-changing configuration values.
var config = struct {
	sync.RWMutex
	m map[string]string
}{
	m: make(map[string]string),
}

// A list of config keys that should be created by initConfig.
// (Any configuration keys should be listed here.)
var configKeys = []string{
	"GerritUsername",
	"GerritPassword",
}

// configEntity is how config values are represented in the datastore.
type configEntity struct {
	Value string
}

// Config returns the value for the given key
// or the empty string if no such key exists.
func Config(c appengine.Context, key string) string {
	config.RLock()
	v, ok := config.m[key]
	config.RUnlock()
	if ok {
		return v
	}

	config.Lock()
	defer config.Unlock()

	// Lookup might have happened after RUnlock; check again.
	if v, ok := config.m[key]; ok {
		return v
	}

	// Lookup config value in datastore.
	k := datastore.NewKey(c, "Config", key, 0, nil)
	ent := configEntity{}
	if err := datastore.Get(c, k, &ent); err != nil {
		c.Errorf("Get Config: %v", err)
		return ""
	}
	// Don't return or cache the dummy value.
	if ent.Value == configDummy {
		return ""
	}
	config.m[key] = ent.Value
	return ent.Value
}

// initConfig is invoked by the initHandler to create an entity for each key in
// configKeys. This makes it easy to edit the configuration values using the
// Datastore Viewer in the App Engine dashboard.
func initConfig(c appengine.Context) {
	for _, key := range configKeys {
		err := datastore.RunInTransaction(c, func(c appengine.Context) error {
			k := datastore.NewKey(c, "Config", key, 0, nil)
			ent := configEntity{}
			if err := datastore.Get(c, k, &ent); err == nil {
				c.Infof("huh? %v", key)
				return nil
			} else if err != datastore.ErrNoSuchEntity {
				return err
			}
			ent.Value = configDummy
			_, err := datastore.Put(c, k, &ent)
			c.Infof("BLAH BLAH %v", key)
			return err
		}, nil)
		if err != nil {
			c.Errorf("initConfig %v: %v", key, err)
		}
	}
}

const configDummy = "[config value unset]"
