// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package buildlet

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/build"
)

type UserPass struct {
	Username string // "user-$USER"
	Password string // buildlet key
}

// A CoordinatorClient makes calls to the build coordinator.
type CoordinatorClient struct {
	// Auth specifies how to authenticate to the coordinator.
	Auth UserPass

	// Instance optionally specifies the build coordinator to connect
	// to. If zero, the production coordinator is used.
	Instance build.CoordinatorInstance

	mu sync.Mutex
	hc *http.Client
}

func (cc *CoordinatorClient) instance() build.CoordinatorInstance {
	if cc.Instance == "" {
		return build.ProdCoordinator
	}
	return cc.Instance
}

func (cc *CoordinatorClient) client() (*http.Client, error) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	if cc.hc != nil {
		return cc.hc, nil
	}
	cc.hc = &http.Client{
		Transport: &http.Transport{
			Dial:    defaultDialer(),
			DialTLS: cc.instance().TLSDialer(),
		},
	}
	return cc.hc, nil
}

// CreateBuildlet creates a new buildlet of the given type on cc.
// It may expire at any time.
// To release it, call Client.Destroy.
func (cc *CoordinatorClient) CreateBuildlet(buildletType string) (*Client, error) {
	hc, err := cc.client()
	if err != nil {
		return nil, err
	}
	ipPort, _ := cc.instance().TLSHostPort() // must succeed if client did
	form := url.Values{
		"type": {buildletType},
	}
	req, _ := http.NewRequest("POST",
		"https://"+ipPort+"/buildlet/create",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(cc.Auth.Username, cc.Auth.Password)
	// TODO: accept a context for deadline/cancelation
	res, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		slurp, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("%s: %s", res.Status, slurp)
	}
	var rb RemoteBuildlet
	if err := json.NewDecoder(res.Body).Decode(&rb); err != nil {
		return nil, err
	}
	if rb.Name == "" {
		return nil, errors.New("buildlet: failed to create remote buildlet; unexpected missing name in response")
	}
	c, err := cc.NamedBuildlet(rb.Name)
	if err != nil {
		return nil, err
	}
	return c, nil
}

type RemoteBuildlet struct {
	Type    string // "openbsd-386"
	Name    string // "buildlet-adg-openbsd-386-2"
	Created time.Time
	Expires time.Time
}

func (cc *CoordinatorClient) RemoteBuildlets() ([]RemoteBuildlet, error) {
	hc, err := cc.client()
	if err != nil {
		return nil, err
	}
	ipPort, _ := cc.instance().TLSHostPort() // must succeed if client did
	req, _ := http.NewRequest("GET", "https://"+ipPort+"/buildlet/list", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(cc.Auth.Username, cc.Auth.Password)
	res, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		slurp, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("%s: %s", res.Status, slurp)
	}
	var ret []RemoteBuildlet
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// NamedBuildlet returns a buildlet client for the named remote buildlet.
// Names are not validated. Use Client.Status to check whether the client works.
func (cc *CoordinatorClient) NamedBuildlet(name string) (*Client, error) {
	hc, err := cc.client()
	if err != nil {
		return nil, err
	}
	ipPort, _ := cc.instance().TLSHostPort() // must succeed if client did
	c := &Client{
		baseURL:        "https://" + ipPort,
		remoteBuildlet: name,
		httpClient:     hc,
		authUser:       cc.Auth.Username,
		password:       cc.Auth.Password,
	}
	c.setCommon()
	return c, nil
}
