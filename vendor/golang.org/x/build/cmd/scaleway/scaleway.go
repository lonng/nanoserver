// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The scaleway command creates ARM servers on Scaleway.com.
package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	token   = flag.String("token", "", "API token")
	org     = flag.String("org", "1f34701d-668b-441b-bf08-0b13544e99de", "Organization ID (default is bradfitz@golang.org's account)")
	image   = flag.String("image", "bebe2c6f-bbb5-4182-9cce-04cab2f44b2b", "Disk image ID; default is the snapshot we made last")
	num     = flag.Int("n", 0, "Number of servers to create; if zero, defaults to a value as a function of --staging")
	tags    = flag.String("tags", "", "Comma-separated list of tags. The build key tags should be of the form 'buildkey_linux-arm_HEXHEXHEXHEXHEX'. If empty, it's automatic.")
	staging = flag.Bool("staging", false, "If true, deploy staging instances (with staging names and tags) instead of prod.")
)

func main() {
	flag.Parse()
	if *tags == "" {
		if *staging {
			*tags = defaultBuilderTags("gobuilder-staging.key")
		} else {
			*tags = defaultBuilderTags("gobuilder-master.key")
		}
	}
	if *num == 0 {
		if *staging {
			*num = 5
		} else {
			*num = 20
		}
	}
	if *token == "" {
		file := filepath.Join(os.Getenv("HOME"), "keys/go-scaleway.token")
		slurp, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("No --token specified and error reading backup token file: %v", err)
		}
		*token = strings.TrimSpace(string(slurp))
	}

	cl := &Client{Token: *token}
	serverList, err := cl.Servers()
	if err != nil {
		log.Fatal(err)
	}
	servers := map[string]*Server{}
	for _, s := range serverList {
		servers[s.Name] = s
	}

	for i := 1; i <= *num; i++ {
		name := fmt.Sprintf("scaleway-prod-%02d", i)
		if *staging {
			name = fmt.Sprintf("scaleway-staging-%02d", i)
		}
		_, ok := servers[name]
		if !ok {
			tags := strings.Split(*tags, ",")
			if *staging {
				tags = append(tags, "staging")
			}
			body, err := json.Marshal(createServerRequest{
				Org:   *org,
				Name:  name,
				Image: *image,
				Tags:  tags,
			})
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Doing req %q for token %q", body, *token)
			req, err := http.NewRequest("POST", "https://api.scaleway.com/servers", bytes.NewReader(body))
			if err != nil {
				log.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Auth-Token", *token)
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Create of %v: %v", i, res.Status)
			res.Body.Close()
		}
	}

	serverList, err = cl.Servers()
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range serverList {
		if strings.HasSuffix(s.Name, "-prep") || strings.HasSuffix(s.Name, "-hand") {
			continue
		}
		if s.State == "stopped" {
			log.Printf("Powering on %s = %v", s.ID, cl.PowerOn(s.ID))
		}
	}
}

type createServerRequest struct {
	Org   string   `json:"organization"`
	Name  string   `json:"name"`
	Image string   `json:"image"`
	Tags  []string `json:"tags"`
}

type Client struct {
	Token string
}

func (c *Client) PowerOn(serverID string) error {
	return c.serverAction(serverID, "poweron")
}

func (c *Client) serverAction(serverID, action string) error {
	req, _ := http.NewRequest("POST", "https://api.scaleway.com/servers/"+serverID+"/action", strings.NewReader(fmt.Sprintf(`{"action":"%s"}`, action)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", c.Token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return fmt.Errorf("Error doing %q on %s: %v", action, serverID, res.Status)
	}
	return nil
}

func (c *Client) Servers() ([]*Server, error) {
	req, _ := http.NewRequest("GET", "https://api.scaleway.com/servers", nil)
	req.Header.Set("X-Auth-Token", c.Token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Failed to get Server list: %v", res.Status)
	}
	var jres struct {
		Servers []*Server `json:"servers"`
	}
	err = json.NewDecoder(res.Body).Decode(&jres)
	return jres.Servers, err
}

type Server struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	PublicIP  *IP      `json:"public_ip"`
	PrivateIP string   `json:"private_ip"`
	Tags      []string `json:"tags"`
	State     string   `json:"state"`
	Image     *Image   `json:"image"`
}

type Image struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type IP struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

// defaultBuilderTags returns the default value of the "tags" flag.
// It returns a comma-separated list of builder tags (each of the form buildkey_$(BUILDER)_$(SECRETHEX)).
func defaultBuilderTags(baseKeyFile string) string {
	keyFile := filepath.Join(os.Getenv("HOME"), "keys", baseKeyFile)
	slurp, err := ioutil.ReadFile(keyFile)
	if err != nil {
		log.Fatal(err)
	}
	var tags []string
	for _, builder := range []string{"linux-arm", "linux-arm-arm5"} {
		h := hmac.New(md5.New, bytes.TrimSpace(slurp))
		h.Write([]byte(builder))
		tags = append(tags, fmt.Sprintf("buildkey_%s_%x", builder, h.Sum(nil)))
	}
	return strings.Join(tags, ",")
}
