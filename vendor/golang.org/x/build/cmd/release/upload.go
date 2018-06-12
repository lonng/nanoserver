// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
)

const (
	uploadURL     = "https://golang.org/dl/upload"
	projectID     = "999119582588"
	storageBucket = "golang"
)

// File represents a file on the golang.org downloads page.
// It should be kept in sync with the download code in x/tools/godoc/dl.
type File struct {
	Filename       string
	OS             string
	Arch           string
	Version        string
	ChecksumSHA256 string
	Size           int64
	Kind           string // "archive", "installer", "source"
}

// fileRe matches the files created by the release tool, such as:
//   go1.5beta2.src.tar.gz
//   go1.5.1.linux-386.tar.gz
//   go1.5.windows-amd64.msi
var fileRe = regexp.MustCompile(`^(go[a-z0-9-.]+)\.(src|([a-z0-9]+)-([a-z0-9]+)(?:-([a-z0-9.]+))?)\.(tar\.gz|zip|pkg|msi)$`)

func upload(files []string) error {
	ctx := context.Background()
	c, err := storageClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	for _, name := range files {
		base := filepath.Base(name)
		log.Printf("Uploading %v", base)
		m := fileRe.FindStringSubmatch(base)
		if m == nil {
			return fmt.Errorf("unrecognized file: %q", base)
		}
		var b Build
		version := m[1]
		if m[2] == "src" {
			b.Source = true
		} else {
			b.OS = m[3]
			b.Arch = m[4]
		}
		if err := uploadFile(ctx, c, &b, version, name); err != nil {
			return err
		}
	}
	return nil
}

func uploadFile(ctx context.Context, c *storage.Client, b *Build, version, filename string) error {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	base := filepath.Base(filename)
	checksum := fmt.Sprintf("%x", sha256.Sum256(file))

	// Upload file to Google Cloud Storage.
	if err := putObject(ctx, c, base, file); err != nil {
		return fmt.Errorf("uploading %q: %v", base, err)
	}
	// Upload file.sha256.
	if err := putObject(ctx, c, base+".sha256", []byte(checksum)); err != nil {
		return fmt.Errorf("uploading %q: %v", base+".sha256", err)
	}

	// Post file details to golang.org.
	var kind string
	switch {
	case b.Source:
		kind = "source"
	case strings.HasSuffix(base, ".tar.gz"), strings.HasSuffix(base, ".zip"):
		kind = "archive"
	case strings.HasSuffix(base, ".msi"), strings.HasSuffix(base, ".pkg"):
		kind = "installer"
	}
	req, err := json.Marshal(File{
		Filename:       base,
		Version:        version,
		OS:             b.OS,
		Arch:           b.Arch,
		ChecksumSHA256: checksum,
		Size:           int64(len(file)),
		Kind:           kind,
	})
	if err != nil {
		return err
	}
	v := url.Values{"user": {*user}, "key": []string{userToken()}}
	u := fmt.Sprintf("%s?%s", uploadURL, v.Encode())
	resp, err := http.Post(u, "application/json", bytes.NewReader(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %v\n%s", resp.Status, b)
	}
	return nil

}

func putObject(ctx context.Context, c *storage.Client, name string, body []byte) error {
	wr := c.Bucket(storageBucket).Object(name).NewWriter(ctx)
	wr.ACL = []storage.ACLRule{
		{Entity: storage.AllUsers, Role: storage.RoleReader},
		// If you don't give the owners access, the web UI seems to
		// have a bug and doesn't have access to see that it's public,
		// so won't render the "Shared Publicly" link. So we do that,
		// even though it's dumb and unnecessary otherwise:
		{Entity: storage.ACLEntity("project-owners-" + projectID), Role: storage.RoleOwner},
	}
	if _, err := wr.Write(body); err != nil {
		return err
	}
	return wr.Close()
}

func storageClient(ctx context.Context) (*storage.Client, error) {
	file := filepath.Join(os.Getenv("HOME"), "keys", "golang-org.service.json")
	blob, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	config, err := google.JWTConfigFromJSON(blob, storage.ScopeReadWrite)
	if err != nil {
		return nil, err
	}
	return storage.NewClient(ctx, cloud.WithBaseHTTP(config.Client(ctx)))
}
