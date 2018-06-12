// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"time"

	"google.golang.org/cloud/datastore"

	"golang.org/x/net/context"
)

// Process is a datastore record about the lifetime of a coordinator process.
//
// Example GQL query:
// SELECT * From Process where LastHeartbeat > datetime("2016-01-01T00:00:00Z")
type ProcessRecord struct {
	ID            string
	Start         time.Time
	LastHeartbeat time.Time

	// TODO: version, who deployed, CoreOS version, Docker version,
	// GCE instance type?
}

func updateInstanceRecord() {
	if dsClient == nil {
		return
	}
	ctx := context.Background()
	for {
		key := datastore.NewKey(ctx, "Process", processID, 0, nil)
		_, err := dsClient.Put(ctx, key, &ProcessRecord{
			ID:            processID,
			Start:         processStartTime,
			LastHeartbeat: time.Now(),
		})
		if err != nil {
			log.Printf("datastore Process Put: %v", err)
		}
		time.Sleep(30 * time.Second)
	}
}

// BuildRecord is the datastore entity we write both at the beginning
// and end of a build. Some fields are not updated until the end.
type BuildRecord struct {
	ID        string
	ProcessID string
	StartTime time.Time
	IsTry     bool // is trybot run
	GoRev     string
	Rev       string // same as GoRev for repo "go"
	Repo      string // "go", "net", etc.
	Builder   string // "linux-amd64-foo"
	OS        string // "linux"
	Arch      string // "amd64"

	EndTime    time.Time
	Seconds    float64
	Result     string // empty string, "ok", "fail"
	FailureURL string `datastore:",noindex"`

	// TODO(bradfitz): log which reverse buildlet we got?
	// Buildlet string
}

func (br *BuildRecord) put() {
	if dsClient == nil {
		return
	}
	ctx := context.Background()
	key := datastore.NewKey(ctx, "Build", br.ID, 0, nil)
	if _, err := dsClient.Put(ctx, key, br); err != nil {
		log.Printf("datastore Build Put: %v", err)
	}
}
