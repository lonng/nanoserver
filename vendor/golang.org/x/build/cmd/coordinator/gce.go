// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Code interacting with Google Compute Engine (GCE) and
// a GCE implementation of the BuildletPool interface.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/build/buildenv"
	"golang.org/x/build/buildlet"
	"golang.org/x/build/dashboard"
	"golang.org/x/build/gerrit"
	"golang.org/x/build/internal/lru"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	monitoring "google.golang.org/api/cloudmonitoring/v2beta2"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/cloud"
	"google.golang.org/cloud/compute/metadata"
	"google.golang.org/cloud/datastore"
	"google.golang.org/cloud/storage"
)

func init() {
	buildlet.GCEGate = gceAPIGate
}

// apiCallTicker ticks regularly, preventing us from accidentally making
// GCE API calls too quickly. Our quota is 20 QPS, but we temporarily
// limit ourselves to less than that.
var apiCallTicker = time.NewTicker(time.Second / 10)

func gceAPIGate() {
	<-apiCallTicker.C
}

// Initialized by initGCE:
var (
	buildEnv *buildenv.Environment

	dsClient       *datastore.Client
	computeService *compute.Service
	tokenSource    oauth2.TokenSource
	serviceCtx     context.Context
	errTryDeps     error // non-nil if try bots are disabled
	gerritClient   *gerrit.Client
	storageClient  *storage.Client
	inStaging      bool // are we running in the staging project? (named -dev)

	initGCECalled bool
)

func initGCE() error {
	initGCECalled = true
	var err error

	// If the coordinator is running on a GCE instance and a
	// buildEnv was not specified with the env flag, set the
	// buildEnvName to the project ID
	if *buildEnvName == "" {
		if *mode == "dev" {
			*buildEnvName = "dev"
		} else if metadata.OnGCE() {
			*buildEnvName, err = metadata.ProjectID()
			if err != nil {
				log.Fatalf("metadata.ProjectID: %v", err)
			}
		}
	}

	buildEnv = buildenv.ByProjectID(*buildEnvName)
	inStaging = (buildEnv == buildenv.Staging)

	// If running on GCE, override the zone and static IP, and check service account permissions.
	if metadata.OnGCE() {
		projectZone, err := metadata.Get("instance/zone")
		if err != nil || projectZone == "" {
			return fmt.Errorf("failed to get current GCE zone: %v", err)
		}
		// Convert the zone from "projects/1234/zones/us-central1-a" to "us-central1-a".
		projectZone = path.Base(projectZone)
		buildEnv.Zone = projectZone

		buildEnv.StaticIP, err = metadata.ExternalIP()
		if err != nil {
			return fmt.Errorf("ExternalIP: %v", err)
		}

		if !hasComputeScope() {
			return errors.New("The coordinator is not running with access to read and write Compute resources. VM support disabled.")

		}
	}

	cfgDump, _ := json.MarshalIndent(buildEnv, "", "  ")
	log.Printf("Loaded configuration %q for project %q:\n%s", *buildEnvName, buildEnv.ProjectName, cfgDump)

	tokenSource, err = google.DefaultTokenSource(oauth2.NoContext, compute.CloudPlatformScope, monitoring.MonitoringScope)
	if err != nil {
		log.Fatalf("failed to get a token source: %v", err)
	}
	httpClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	if *mode != "dev" {
		serviceCtx = cloud.NewContext(buildEnv.ProjectName, httpClient)
		storageClient, err = storage.NewClient(serviceCtx, cloud.WithBaseHTTP(httpClient))
		if err != nil {
			log.Fatalf("storage.NewClient: %v", err)
		}
	}

	dsClient, err = datastore.NewClient(context.Background(), buildEnv.ProjectName)
	if err != nil {
		// TODO(bradfitz): make fatal later, once working.
		log.Printf("Error creating datastore client: %v", err)
	}

	computeService, _ = compute.New(httpClient)
	errTryDeps = checkTryBuildDeps()
	if errTryDeps != nil {
		log.Printf("TryBot builders disabled due to error: %v", errTryDeps)
	} else {
		log.Printf("TryBot builders enabled.")
	}

	go gcePool.pollQuotaLoop()
	return nil
}

func checkTryBuildDeps() error {
	if !hasStorageScope() {
		return errors.New("coordinator's GCE instance lacks the storage service scope")
	}
	if *mode == "dev" {
		return errors.New("running in dev mode")
	}
	wr := storageClient.Bucket(buildEnv.LogBucket).Object("hello.txt").NewWriter(serviceCtx)
	fmt.Fprintf(wr, "Hello, world! Coordinator start-up at %v", time.Now())
	if err := wr.Close(); err != nil {
		return fmt.Errorf("test write of a GCS object to bucket %q failed: %v", buildEnv.LogBucket, err)
	}
	if inStaging {
		// Don't expect to write to Gerrit in staging mode.
		gerritClient = gerrit.NewClient("https://go-review.googlesource.com", gerrit.NoAuth)
	} else {
		gobotPass, err := metadata.ProjectAttributeValue("gobot-password")
		if err != nil {
			return fmt.Errorf("failed to get project metadata 'gobot-password': %v", err)
		}
		gerritClient = gerrit.NewClient("https://go-review.googlesource.com",
			gerrit.BasicAuth("git-gobot.golang.org", strings.TrimSpace(string(gobotPass))))
	}

	return nil
}

var gcePool = &gceBuildletPool{}

var _ BuildletPool = (*gceBuildletPool)(nil)

// maxInstances is a temporary hack because we can't get buildlets to boot
// without IPs, and we only have 200 IP addresses.
// TODO(bradfitz): remove this once fixed.
const maxInstances = 190

type gceBuildletPool struct {
	mu sync.Mutex // guards all following

	disabled bool

	// CPU quota usage & limits.
	cpuLeft   int // dead-reckoning CPUs remain
	instLeft  int // dead-reckoning instances remain
	instUsage int
	cpuUsage  int
	addrUsage int
	inst      map[string]time.Time // GCE VM instance name -> creationTime
}

func (p *gceBuildletPool) pollQuotaLoop() {
	if computeService == nil {
		log.Printf("pollQuotaLoop: no GCE access; not checking quota.")
		return
	}
	if buildEnv.ProjectName == "" {
		log.Printf("pollQuotaLoop: no GCE project name confingured; not checking quota.")
		return
	}
	for {
		p.pollQuota()
		time.Sleep(5 * time.Second)
	}
}

func (p *gceBuildletPool) pollQuota() {
	gceAPIGate()
	reg, err := computeService.Regions.Get(buildEnv.ProjectName, buildEnv.Region()).Do()
	if err != nil {
		log.Printf("Failed to get quota for %s/%s: %v", buildEnv.ProjectName, buildEnv.Region(), err)
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, quota := range reg.Quotas {
		switch quota.Metric {
		case "CPUS":
			p.cpuLeft = int(quota.Limit) - int(quota.Usage)
			p.cpuUsage = int(quota.Usage)
		case "INSTANCES":
			p.instLeft = int(quota.Limit) - int(quota.Usage)
			p.instUsage = int(quota.Usage)
		case "IN_USE_ADDRESSES":
			p.addrUsage = int(quota.Usage)
		}
	}
}

func (p *gceBuildletPool) SetEnabled(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.disabled = !enabled
}

func (p *gceBuildletPool) GetBuildlet(ctx context.Context, typ string, lg logger) (bc *buildlet.Client, err error) {
	conf, ok := dashboard.Builders[typ]
	if !ok {
		return nil, fmt.Errorf("gcepool: unknown buildlet type %q", typ)
	}
	qsp := lg.createSpan("awaiting_gce_quota")
	err = p.awaitVMCountQuota(ctx, conf.GCENumCPU())
	qsp.done(err)
	if err != nil {
		return nil, err
	}

	deleteIn, ok := ctx.Value(buildletTimeoutOpt{}).(time.Duration)
	if !ok {
		deleteIn = vmDeleteTimeout
	}

	instName := "buildlet-" + typ + "-rn" + randHex(7)
	p.setInstanceUsed(instName, true)

	gceBuildletSpan := lg.createSpan("create_gce_buildlet", instName)
	defer func() { gceBuildletSpan.done(err) }()

	var (
		needDelete   bool
		createSpan   eventSpan    = lg.createSpan("create_gce_instance", instName)
		waitBuildlet eventSpan    // made after create is done
		curSpan      = createSpan // either instSpan or waitBuildlet
	)

	log.Printf("Creating GCE VM %q for %s", instName, typ)
	bc, err = buildlet.StartNewVM(tokenSource, instName, typ, buildlet.VMOpts{
		ProjectID:   buildEnv.ProjectName,
		Zone:        buildEnv.Zone,
		Description: fmt.Sprintf("Go Builder for %s", typ),
		DeleteIn:    deleteIn,
		OnInstanceRequested: func() {
			log.Printf("GCE VM %q now booting", instName)
		},
		FallbackToFullPrice: func() string {
			lg.logEventTime("gce_fallback_to_full_price", "for "+instName)
			p.setInstanceUsed(instName, false)
			newName := instName + "-f"
			log.Printf("Gave up on preemptible %q; now booting %q", instName, newName)
			instName = newName
			p.setInstanceUsed(instName, true)
			return newName
		},
		OnInstanceCreated: func() {
			needDelete = true

			createSpan.done(nil)
			waitBuildlet = lg.createSpan("wait_buildlet_start", instName)
			curSpan = waitBuildlet
		},
		OnGotInstanceInfo: func() {
			lg.logEventTime("got_instance_info", "waiting_for_buildlet...")
		},
	})
	if err != nil {
		curSpan.done(err)
		log.Printf("Failed to create VM for %s: %v", typ, err)
		if needDelete {
			deleteVM(buildEnv.Zone, instName)
			p.putVMCountQuota(conf.GCENumCPU())
		}
		p.setInstanceUsed(instName, false)
		return nil, err
	}
	waitBuildlet.done(nil)
	bc.SetDescription("GCE VM: " + instName)
	bc.SetOnHeartbeatFailure(func() {
		p.putBuildlet(bc, typ, instName)
	})
	return bc, nil
}

func (p *gceBuildletPool) putBuildlet(bc *buildlet.Client, typ, instName string) error {
	// TODO(bradfitz): add the buildlet to a freelist (of max N
	// items) for up to 10 minutes since when it got started if
	// it's never seen a command execution failure, and we can
	// wipe all its disk content? (perhaps wipe its disk content
	// when it's retrieved, like the reverse buildlet pool) But
	// this will require re-introducing a distinction in the
	// buildlet client library between Close, Destroy/Halt, and
	// tracking execution errors.  That was all half-baked before
	// and thus removed. Now Close always destroys everything.
	deleteVM(buildEnv.Zone, instName)
	p.setInstanceUsed(instName, false)

	conf, ok := dashboard.Builders[typ]
	if !ok {
		panic("failed to lookup conf") // should've worked if we did it before
	}
	p.putVMCountQuota(conf.GCENumCPU())
	return nil
}

func (p *gceBuildletPool) WriteHTMLStatus(w io.Writer) {
	fmt.Fprintf(w, "<b>GCE pool</b> capacity: %s", p.capacityString())
	const show = 6 // must be even
	active := p.instancesActive()
	if len(active) > 0 {
		fmt.Fprintf(w, "<ul>")
		for i, inst := range active {
			if i < show/2 || i >= len(active)-(show/2) {
				fmt.Fprintf(w, "<li>%v, %v</li>\n", inst.name, time.Since(inst.creation))
			} else if i == show/2 {
				fmt.Fprintf(w, "<li>... %d of %d total omitted ...</li>\n", len(active)-show, len(active))
			}
		}
		fmt.Fprintf(w, "</ul>")
	}
}

func (p *gceBuildletPool) String() string {
	return fmt.Sprintf("GCE pool capacity: %s", p.capacityString())
}

func (p *gceBuildletPool) capacityString() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return fmt.Sprintf("%d/%d instances; %d/%d CPUs",
		len(p.inst), p.instUsage+p.instLeft,
		p.cpuUsage, p.cpuUsage+p.cpuLeft)
}

// awaitVMCountQuota waits for numCPU CPUs of quota to become available,
// or returns ctx.Err.
func (p *gceBuildletPool) awaitVMCountQuota(ctx context.Context, numCPU int) error {
	// Poll every 2 seconds, which could be better, but works and
	// is simple.
	for {
		if p.tryAllocateQuota(numCPU) {
			return nil
		}
		select {
		case <-time.After(2 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *gceBuildletPool) tryAllocateQuota(numCPU int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.disabled {
		return false
	}
	if p.cpuLeft >= numCPU && p.instLeft >= 1 && len(p.inst) < maxInstances && p.addrUsage < maxInstances {
		p.cpuUsage += numCPU
		p.cpuLeft -= numCPU
		p.instLeft--
		p.addrUsage++
		return true
	}
	return false
}

// putVMCountQuota adjusts the dead-reckoning of our quota usage by
// one instance and cpu CPUs.
func (p *gceBuildletPool) putVMCountQuota(cpu int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cpuUsage -= cpu
	p.cpuLeft += cpu
	p.instLeft++
}

func (p *gceBuildletPool) setInstanceUsed(instName string, used bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.inst == nil {
		p.inst = make(map[string]time.Time)
	}
	if used {
		p.inst[instName] = time.Now()
	} else {
		delete(p.inst, instName)
	}
}

func (p *gceBuildletPool) instanceUsed(instName string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.inst[instName]
	return ok
}

func (p *gceBuildletPool) instancesActive() (ret []resourceTime) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for name, create := range p.inst {
		ret = append(ret, resourceTime{
			name:     name,
			creation: create,
		})
	}
	sort.Sort(byCreationTime(ret))
	return ret
}

// resourceTime is a GCE instance or Kube pod name and its creation time.
type resourceTime struct {
	name     string
	creation time.Time
}

type byCreationTime []resourceTime

func (s byCreationTime) Len() int           { return len(s) }
func (s byCreationTime) Less(i, j int) bool { return s[i].creation.Before(s[j].creation) }
func (s byCreationTime) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// cleanUpOldVMs loops forever and periodically enumerates virtual
// machines and deletes those which have expired.
//
// A VM is considered expired if it has a "delete-at" metadata
// attribute having a unix timestamp before the current time.
//
// This is the safety mechanism to delete VMs which stray from the
// normal deleting process. VMs are created to run a single build and
// should be shut down by a controlling process. Due to various types
// of failures, they might get stranded. To prevent them from getting
// stranded and wasting resources forever, we instead set the
// "delete-at" metadata attribute on them when created to some time
// that's well beyond their expected lifetime.
func (p *gceBuildletPool) cleanUpOldVMs() {
	if *mode == "dev" {
		return
	}
	if computeService == nil {
		return
	}

	// TODO(bradfitz): remove this list and just query it from the compute API?
	// http://godoc.org/google.golang.org/api/compute/v1#RegionsService.Get
	// and Region.Zones: http://godoc.org/google.golang.org/api/compute/v1#Region

	for {
		for _, zone := range buildEnv.ZonesToClean {
			if err := p.cleanZoneVMs(zone); err != nil {
				log.Printf("Error cleaning VMs in zone %q: %v", zone, err)
			}
		}
		time.Sleep(time.Minute)
	}
}

// cleanZoneVMs is part of cleanUpOldVMs, operating on a single zone.
func (p *gceBuildletPool) cleanZoneVMs(zone string) error {
	// Fetch the first 500 (default) running instances and clean
	// thoes. We expect that we'll be running many fewer than
	// that. Even if we have more, eventually the first 500 will
	// either end or be cleaned, and then the next call will get a
	// partially-different 500.
	// TODO(bradfitz): revist this code if we ever start running
	// thousands of VMs.
	gceAPIGate()
	list, err := computeService.Instances.List(buildEnv.ProjectName, zone).Do()
	if err != nil {
		return fmt.Errorf("listing instances: %v", err)
	}
	for _, inst := range list.Items {
		if inst.Metadata == nil {
			// Defensive. Not seen in practice.
			continue
		}
		sawDeleteAt := false
		for _, it := range inst.Metadata.Items {
			if it.Key == "delete-at" {
				sawDeleteAt = true
				if it.Value == nil {
					log.Printf("missing delete-at value; ignoring")
					continue
				}
				unixDeadline, err := strconv.ParseInt(*it.Value, 10, 64)
				if err != nil {
					log.Printf("invalid delete-at value %q seen; ignoring", it.Value)
				}
				if err == nil && time.Now().Unix() > unixDeadline {
					log.Printf("Deleting expired VM %q in zone %q ...", inst.Name, zone)
					deleteVM(zone, inst.Name)
				}
			}
		}
		// Delete buildlets (things we made) from previous
		// generations. Only deleting things starting with "buildlet-"
		// is a historical restriction, but still fine for paranoia.
		if sawDeleteAt && strings.HasPrefix(inst.Name, "buildlet-") && !p.instanceUsed(inst.Name) {
			if _, ok := deletedVMCache.Get(inst.Name); !ok {
				log.Printf("Deleting VM %q in zone %q from an earlier coordinator generation ...", inst.Name, zone)
				deleteVM(zone, inst.Name)
			}
		}
	}
	return nil
}

var deletedVMCache = lru.New(100) // keyed by instName

// deleteVM starts a delete of an instance in a given zone.
//
// It either returns an operation name (if delete is pending) or the
// empty string if the instance didn't exist.
func deleteVM(zone, instName string) (operation string, err error) {
	deletedVMCache.Add(instName, token{})
	gceAPIGate()
	op, err := computeService.Instances.Delete(buildEnv.ProjectName, zone, instName).Do()
	apiErr, ok := err.(*googleapi.Error)
	if ok {
		if apiErr.Code == 404 {
			return "", nil
		}
	}
	if err != nil {
		log.Printf("Failed to delete instance %q in zone %q: %v", instName, zone, err)
		return "", err
	}
	log.Printf("Sent request to delete instance %q in zone %q. Operation ID, Name: %v, %v", instName, zone, op.Id, op.Name)
	return op.Name, nil
}

func hasScope(want string) bool {
	// If not on GCE, assume full access
	if !metadata.OnGCE() {
		return true
	}
	scopes, err := metadata.Scopes("default")
	if err != nil {
		log.Printf("failed to query metadata default scopes: %v", err)
		return false
	}
	for _, v := range scopes {
		if v == want {
			return true
		}
	}
	return false
}

func hasComputeScope() bool {
	return hasScope(compute.ComputeScope) || hasScope(compute.CloudPlatformScope)
}

func hasStorageScope() bool {
	return hasScope(storage.ScopeReadWrite) || hasScope(storage.ScopeFullControl) || hasScope(compute.CloudPlatformScope)
}
