// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

/*
This file implements reverse buildlets. These are buildlets that are not
started by the coordinator. They dial the coordinator and then accept
instructions. This feature is used for machines that cannot be started by
an API, for example real OS X machines with iOS and Android devices attached.

You can test this setup locally. In one terminal start a coordinator.
It will default to dev mode, using a dummy TLS cert and not talking to GCE.

	$ coordinator

In another terminal, start a reverse buildlet:

	$ buildlet -reverse "darwin-amd64"

It will dial and register itself with the coordinator. To confirm the
coordinator can see the buildlet, check the logs output or visit its
diagnostics page: https://localhost:8119. To send the buildlet some
work, go to:

	https://localhost:8119/dosomework
*/

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/build/buildlet"
	"golang.org/x/build/revdial"
	"golang.org/x/net/context"
)

const minBuildletVersion = 1

var reversePool = &reverseBuildletPool{
	available: make(chan token, 1),
}

type token struct{}

type reverseBuildletPool struct {
	available chan token // best-effort tickle when any buildlet becomes free

	mu        sync.Mutex // guards buildlets and their fields
	buildlets []*reverseBuildlet
}

var errInUse = errors.New("all buildlets are in use")

func (p *reverseBuildletPool) tryToGrab(machineType string) (*buildlet.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	usableCount := 0
	for _, b := range p.buildlets {
		usable := false
		for _, m := range b.modes {
			if m == machineType {
				usable = true
				usableCount++
				break
			}
		}
		if usable && b.inUseAs == "" {
			// Found an unused match.
			b.inUseAs = machineType
			b.inUseTime = time.Now()
			return b.client, nil
		}
	}
	if usableCount == 0 {
		return nil, fmt.Errorf("no buildlets registered for machine type %q", machineType)
	}
	return nil, errInUse
}

func (p *reverseBuildletPool) noteBuildletAvailable() {
	select {
	case p.available <- token{}:
	default:
	}
}

// nukeBuildlet wipes out victim as a buildlet we'll ever return again,
// and closes its TCP connection in hopes that it will fix itself
// later.
func (p *reverseBuildletPool) nukeBuildlet(victim *buildlet.Client) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, rb := range p.buildlets {
		if rb.client == victim {
			defer rb.conn.Close()
			p.buildlets = append(p.buildlets[:i], p.buildlets[i+1:]...)
			return
		}
	}
}

// healthCheckBuildletLoop periodically requests the status from b.
// If the buildlet fails to respond promptly, it is removed from the pool.
func (p *reverseBuildletPool) healthCheckBuildletLoop(b *reverseBuildlet) {
	for {
		time.Sleep(time.Duration(10+rand.Intn(5)) * time.Second)
		if !p.healthCheckBuildlet(b) {
			return
		}
	}
}

func (p *reverseBuildletPool) healthCheckBuildlet(b *reverseBuildlet) bool {
	p.mu.Lock()
	if b.inUseAs == "health" { // sanity check
		panic("previous health check still running")
	}
	if b.inUseAs != "" {
		p.mu.Unlock()
		return true // skip busy buildlets
	}
	b.inUseAs = "health"
	b.inUseTime = time.Now()
	res := make(chan error, 1)
	go func() {
		_, err := b.client.Status()
		res <- err
	}()
	p.mu.Unlock()

	t := time.NewTimer(5 * time.Second) // give buildlets time to respond
	var err error
	select {
	case err = <-res:
		t.Stop()
	case <-t.C:
		err = errors.New("health check timeout")
	}

	if err != nil {
		// remove bad buildlet
		log.Printf("Health check fail; removing reverse buildlet %s %v: %v", b.client, b.modes, err)
		go b.client.Close()
		go p.nukeBuildlet(b.client)
		return false
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if b.inUseAs != "health" {
		// buildlet was grabbed while lock was released; harmless.
		return true
	}
	b.inUseAs = ""
	b.inUseTime = time.Now()
	p.noteBuildletAvailable()
	return true
}

var (
	highPriorityBuildletMu sync.Mutex
	highPriorityBuildlet   = make(map[string]chan *buildlet.Client)
)

func highPriChan(typ string) chan *buildlet.Client {
	highPriorityBuildletMu.Lock()
	defer highPriorityBuildletMu.Unlock()
	if c, ok := highPriorityBuildlet[typ]; ok {
		return c
	}
	c := make(chan *buildlet.Client)
	highPriorityBuildlet[typ] = c
	return c
}

func (p *reverseBuildletPool) GetBuildlet(ctx context.Context, machineType string, lg logger) (*buildlet.Client, error) {
	seenErrInUse := false
	isHighPriority, _ := ctx.Value(highPriorityOpt{}).(bool)
	sp := lg.createSpan("wait_static_builder", machineType)
	for {
		b, err := p.tryToGrab(machineType)
		if err == errInUse {
			if !seenErrInUse {
				lg.logEventTime("waiting_machine_in_use")
				seenErrInUse = true
			}
			var highPri chan *buildlet.Client
			if isHighPriority {
				highPri = highPriChan(machineType)
			}
			select {
			case <-ctx.Done():
				return nil, sp.done(ctx.Err())
			case bc := <-highPri:
				sp.done(nil)
				return p.cleanedBuildlet(bc, lg)
			// As multiple goroutines can be listening for
			// the available signal, it must be treated as
			// a best effort signal. So periodically try
			// to grab a buildlet again:
			case <-time.After(10 * time.Second):
			case <-p.available:
			}
		} else if err != nil {
			sp.done(err)
			return nil, err
		} else {
			select {
			case highPriChan(machineType) <- b:
				// Somebody else was more important.
			default:
				sp.done(nil)
				return p.cleanedBuildlet(b, lg)
			}
		}
	}
}

func (p *reverseBuildletPool) cleanedBuildlet(b *buildlet.Client, lg logger) (*buildlet.Client, error) {
	// Clean up any files from previous builds.
	sp := lg.createSpan("clean_buildlet", b.String())
	err := b.RemoveAll(".")
	sp.done(err)
	if err != nil {
		b.Close()
		return nil, err
	}
	return b, nil
}

func (p *reverseBuildletPool) WriteHTMLStatus(w io.Writer) {
	// total maps from a builder type to the number of machines which are
	// capable of that role.
	total := make(map[string]int)
	// inUse and inUseOther track the number of machines using machines.
	// inUse is how many machines are building that type, and inUseOther counts
	// how many machines are occupied doing a similar role on that hardware.
	// e.g. "darwin-amd64-10_10" occupied as a "darwin-arm-a5ios",
	// or "linux-arm" as a "linux-arm-arm5" count as inUseOther.
	inUse := make(map[string]int)
	inUseOther := make(map[string]int)

	var machineBuf bytes.Buffer
	p.mu.Lock()
	buildlets := append([]*reverseBuildlet(nil), p.buildlets...)
	sort.Sort(byModeThenHostname(buildlets))
	for _, b := range buildlets {
		machStatus := "<i>idle</i>"
		if b.inUseAs != "" {
			machStatus = "working as <b>" + b.inUseAs + "</b>"
		}
		fmt.Fprintf(&machineBuf, "<li>%s (%s) version %s, %s: connected %v, %s for %v</li>\n",
			b.hostname,
			b.conn.RemoteAddr(),
			b.version,
			strings.Join(b.modes, ", "),
			time.Since(b.regTime),
			machStatus,
			time.Since(b.inUseTime))
		for _, mode := range b.modes {
			if b.inUseAs != "" && b.inUseAs != "health" {
				if mode == b.inUseAs {
					inUse[mode]++
				} else {
					inUseOther[mode]++
				}
			}
			total[mode]++
		}
	}
	p.mu.Unlock()

	var modes []string
	for mode := range total {
		modes = append(modes, mode)
	}
	sort.Strings(modes)

	io.WriteString(w, "<b>Reverse pool summary</b><ul>")
	if len(modes) == 0 {
		io.WriteString(w, "<li>no connections</li>")
	}
	for _, mode := range modes {
		use, other := inUse[mode], inUseOther[mode]
		if use+other == 0 {
			fmt.Fprintf(w, "<li>%s: 0/%d</li>", mode, total[mode])
		} else {
			fmt.Fprintf(w, "<li>%s: %d/%d (%d + %d other)</li>", mode, use+other, total[mode], use, other)
		}
	}
	io.WriteString(w, "</ul>")

	fmt.Fprintf(w, "<b>Reverse pool machine detail</b><ul>%s</ul>", machineBuf.Bytes())
}

func (p *reverseBuildletPool) String() string {
	p.mu.Lock()
	inUse := 0
	total := len(p.buildlets)
	for _, b := range p.buildlets {
		if b.inUseAs != "" && b.inUseAs != "health" {
			inUse++
		}
	}
	p.mu.Unlock()

	return fmt.Sprintf("Reverse pool capacity: %d/%d %s", inUse, total, p.Modes())
}

// Modes returns the a deduplicated list of buildlet modes curently supported
// by the pool. Buildlet modes are described on reverseBuildlet comments.
func (p *reverseBuildletPool) Modes() (modes []string) {
	mm := make(map[string]bool)
	p.mu.Lock()
	for _, b := range p.buildlets {
		for _, mode := range b.modes {
			mm[mode] = true
		}
	}
	p.mu.Unlock()

	for mode := range mm {
		modes = append(modes, mode)
	}
	sort.Strings(modes)
	return modes
}

// CanBuild reports whether the pool has a machine capable of building mode.
// The machine may be in use, so you may have to wait.
func (p *reverseBuildletPool) CanBuild(mode string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, b := range p.buildlets {
		for _, m := range b.modes {
			if m == mode {
				return true
			}
		}
	}
	return false
}

func (p *reverseBuildletPool) addBuildlet(b *reverseBuildlet) {
	p.mu.Lock()
	defer p.noteBuildletAvailable()
	defer p.mu.Unlock()
	p.buildlets = append(p.buildlets, b)
	go p.healthCheckBuildletLoop(b)
}

// reverseBuildlet is a registered reverse buildlet.
// Its immediate fields are guarded by the reverseBuildletPool mutex.
type reverseBuildlet struct {
	// hostname is the name of the buildlet host.
	// It doesn't have to be a complete DNS name.
	hostname string
	// version is the reverse buildlet's version.
	version string

	// sessRand is the unique random number for every unique buildlet session.
	sessRand string

	client  *buildlet.Client
	conn    net.Conn
	regTime time.Time // when it was first connected

	// modes is the set of valid modes for this buildlet.
	//
	// A mode is the equivalent of a builder name, for example
	// "darwin-amd64", "android-arm", or "linux-amd64-race".
	//
	// Each buildlet may potentially have several modes. For example a
	// Mac OS X machine with an attached iOS device may be registered
	// as both "darwin-amd64", "darwin-arm64".
	modes []string

	// inUseAs signifies that the buildlet is in use as the named mode.
	// inUseTime is when it entered that state.
	// Both are guarded by the mutex on reverseBuildletPool.
	inUseAs   string
	inUseTime time.Time
}

func (b *reverseBuildlet) firstMode() string {
	if len(b.modes) == 0 {
		return ""
	}
	return b.modes[0]
}

func handleReverse(w http.ResponseWriter, r *http.Request) {
	if r.TLS == nil {
		http.Error(w, "buildlet registration requires SSL", http.StatusInternalServerError)
		return
	}
	// Check build keys.
	modes := r.Header["X-Go-Builder-Type"]
	gobuildkeys := r.Header["X-Go-Builder-Key"]
	if len(modes) == 0 || len(modes) != len(gobuildkeys) {
		http.Error(w, fmt.Sprintf("need at least one mode and matching key, got %d/%d", len(modes), len(gobuildkeys)), http.StatusPreconditionFailed)
		return
	}
	for i, m := range modes {
		if gobuildkeys[i] != builderKey(m) {
			http.Error(w, fmt.Sprintf("bad key for mode %q", m), http.StatusPreconditionFailed)
			return
		}
	}

	conn, bufrw, err := w.(http.Hijacker).Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	hostname := r.Header.Get("X-Go-Builder-Hostname")

	revDialer := revdial.NewDialer(bufrw, conn)

	log.Printf("Registering reverse buildlet %q (%s) for modes %v", hostname, r.RemoteAddr, modes)

	(&http.Response{StatusCode: http.StatusSwitchingProtocols, Proto: "HTTP/1.1"}).Write(conn)

	client := buildlet.NewClient(hostname, buildlet.NoKeyPair)
	client.SetHTTPClient(&http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return revDialer.Dial()
			},
		},
	})
	client.SetDescription(fmt.Sprintf("reverse peer %s/%s for modes %v", hostname, r.RemoteAddr, modes))

	var isDead struct {
		sync.Mutex
		v bool
	}
	client.SetOnHeartbeatFailure(func() {
		isDead.Lock()
		isDead.v = true
		isDead.Unlock()
		conn.Close()
		reversePool.nukeBuildlet(client)
	})

	// If the reverse dialer (which is always reading from the
	// conn) detects that the remote went away, close the buildlet
	// client proactively show
	go func() {
		<-revDialer.Done()
		isDead.Lock()
		defer isDead.Unlock()
		if !isDead.v {
			client.Close()
		}
	}()
	tstatus := time.Now()
	status, err := client.Status()
	if err != nil {
		log.Printf("Reverse connection %s/%s for modes %v did not answer status after %v: %v",
			hostname, r.RemoteAddr, modes, time.Since(tstatus), err)
		conn.Close()
		return
	}
	if status.Version < minBuildletVersion {
		log.Printf("Buildlet too old: %s, %+v", r.RemoteAddr, status)
		conn.Close()
		return
	}
	log.Printf("Buildlet %s/%s: %+v for %s", hostname, r.RemoteAddr, status, modes)

	now := time.Now()
	b := &reverseBuildlet{
		hostname:  hostname,
		version:   r.Header.Get("X-Go-Builder-Version"),
		modes:     modes,
		client:    client,
		conn:      conn,
		inUseTime: now,
		regTime:   now,
	}
	reversePool.addBuildlet(b)
	registerBuildlet(modes) // testing only
}

var registerBuildlet = func(modes []string) {} // test hook

type byModeThenHostname []*reverseBuildlet

func (s byModeThenHostname) Len() int      { return len(s) }
func (s byModeThenHostname) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byModeThenHostname) Less(i, j int) bool {
	bi, bj := s[i], s[j]
	mi, mj := bi.firstMode(), bj.firstMode()
	if mi == mj {
		return bi.hostname < bj.hostname
	}
	return mi < mj
}
