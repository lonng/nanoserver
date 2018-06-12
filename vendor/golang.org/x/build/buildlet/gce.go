// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package buildlet

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/build/buildenv"
	"golang.org/x/build/dashboard"
	"golang.org/x/oauth2"
	"google.golang.org/api/compute/v1"
)

// GCEGate optionally specifies a function to run before any GCE API call.
// It's intended to be used to bound QPS rate to GCE.
var GCEGate func()

func apiGate() {
	if GCEGate != nil {
		GCEGate()
	}
}

// VMOpts control how new VMs are started.
type VMOpts struct {
	// Zone is the GCE zone to create the VM in. Required.
	Zone string

	// ProjectID is the GCE project ID. Required.
	ProjectID string

	// TLS optionally specifies the TLS keypair to use.
	// If zero, http without auth is used.
	TLS KeyPair

	// Optional description of the VM.
	Description string

	// Optional metadata to put on the instance.
	Meta map[string]string

	// DeleteIn optionally specifies a duration at which

	// to delete the VM.
	DeleteIn time.Duration

	// OnInstanceRequested optionally specifies a hook to run synchronously
	// after the computeService.Instances.Insert call, but before
	// waiting for its operation to proceed.
	OnInstanceRequested func()

	// OnInstanceCreated optionally specifies a hook to run synchronously
	// after the instance operation succeeds.
	OnInstanceCreated func()

	// OnInstanceCreated optionally specifies a hook to run synchronously
	// after the computeService.Instances.Get call.
	OnGotInstanceInfo func()

	// FallbackToFullPrice optionally specifies a hook to return a new
	// GCE instance name if the first one failed to launch
	// as preemptible. (If you use the same name, GCE complains about
	// resources already existing, even if it failed to be created)
	FallbackToFullPrice func() (newInstname string)
}

// StartNewVM boots a new VM on GCE and returns a buildlet client
// configured to speak to it.
func StartNewVM(ts oauth2.TokenSource, instName, builderType string, opts VMOpts) (*Client, error) {
	computeService, _ := compute.New(oauth2.NewClient(oauth2.NoContext, ts))

	conf, ok := dashboard.Builders[builderType]
	if !ok {
		return nil, fmt.Errorf("invalid builder type %q", builderType)
	}

	zone := opts.Zone
	if zone == "" {
		// TODO: automatic? maybe that's not useful.
		// For now just return an error.
		return nil, errors.New("buildlet: missing required Zone option")
	}
	projectID := opts.ProjectID
	if projectID == "" {
		return nil, errors.New("buildlet: missing required ProjectID option")
	}

	usePreempt := false
Try:
	prefix := "https://www.googleapis.com/compute/v1/projects/" + projectID
	machType := prefix + "/zones/" + zone + "/machineTypes/" + conf.MachineType()
	diskType := "https://www.googleapis.com/compute/v1/projects/" + projectID + "/zones/" + zone + "/diskTypes/pd-ssd"
	if conf.RegularDisk {
		diskType = "" // a spinning disk
	}

	// Request an IP address if this is a world-facing buildlet.
	var accessConfigs []*compute.AccessConfig
	// TODO(bradfitz): remove the "true ||" part once we figure out why the buildlet
	// never boots without an IP address. Userspace seems to hang before we get to the buildlet?
	if true || !opts.TLS.IsZero() {
		accessConfigs = []*compute.AccessConfig{
			&compute.AccessConfig{
				Type: "ONE_TO_ONE_NAT",
				Name: "External NAT",
			},
		}
	}

	instance := &compute.Instance{
		Name:        instName,
		Description: opts.Description,
		MachineType: machType,
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				Type:       "PERSISTENT",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskName:    instName,
					SourceImage: "https://www.googleapis.com/compute/v1/projects/" + projectID + "/global/images/" + conf.VMImage,
					DiskType:    diskType,
				},
			},
		},
		Tags: &compute.Tags{
			// Warning: do NOT list "http-server" or "allow-ssh" (our
			// project's custom tag to allow ssh access) here; the
			// buildlet provides full remote code execution.
			// The https-server is authenticated, though.
			Items: []string{"https-server"},
		},
		Metadata: &compute.Metadata{},
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				AccessConfigs: accessConfigs,
				Network:       prefix + "/global/networks/default",
			},
		},
		Scheduling: &compute.Scheduling{
			Preemptible: usePreempt,
		},
	}
	addMeta := func(key, value string) {
		instance.Metadata.Items = append(instance.Metadata.Items, &compute.MetadataItems{
			Key:   key,
			Value: &value,
		})
	}
	// The buildlet-binary-url is the URL of the buildlet binary
	// which the VMs are configured to download at boot and run.
	// This lets us/ update the buildlet more easily than
	// rebuilding the whole VM image.
	addMeta("buildlet-binary-url", conf.BuildletBinaryURL(buildenv.ByProjectID(opts.ProjectID)))
	addMeta("builder-type", builderType)
	if !opts.TLS.IsZero() {
		addMeta("tls-cert", opts.TLS.CertPEM)
		addMeta("tls-key", opts.TLS.KeyPEM)
		addMeta("password", opts.TLS.Password())
	}

	if opts.DeleteIn != 0 {
		// In case the VM gets away from us (generally: if the
		// coordinator dies while a build is running), then we
		// set this attribute of when it should be killed so
		// we can kill it later when the coordinator is
		// restarted. The cleanUpOldVMs goroutine loop handles
		// that killing.
		addMeta("delete-at", fmt.Sprint(time.Now().Add(opts.DeleteIn).Unix()))
	}

	for k, v := range opts.Meta {
		addMeta(k, v)
	}

	apiGate()
	op, err := computeService.Instances.Insert(projectID, zone, instance).Do()
	if err != nil {
		return nil, fmt.Errorf("Failed to create instance: %v", err)
	}
	condRun(opts.OnInstanceRequested)
	createOp := op.Name

	// Wait for instance create operation to succeed.
OpLoop:
	for {
		time.Sleep(2 * time.Second)
		apiGate()
		op, err := computeService.ZoneOperations.Get(projectID, zone, createOp).Do()
		if err != nil {
			return nil, fmt.Errorf("Failed to get op %s: %v", createOp, err)
		}
		switch op.Status {
		case "PENDING", "RUNNING":
			continue
		case "DONE":
			if op.Error != nil {
				for _, operr := range op.Error.Errors {
					log.Printf("failed to create instance %s in zone %s: %v", instName, zone, operr.Code)
					if operr.Code == "ZONE_RESOURCE_POOL_EXHAUSTED" && usePreempt && opts.FallbackToFullPrice != nil {
						oldName := instName
						usePreempt = false
						instName = opts.FallbackToFullPrice()
						log.Printf("buildlet/gce: retrying without preempt with name %q (previously: %q)", instName, oldName)
						goto Try
					}
					// TODO: catch Code=="QUOTA_EXCEEDED" and "Message" and return
					// a known error value/type.
					return nil, fmt.Errorf("Error creating instance: %+v", operr)
				}
				return nil, errors.New("Failed to start.")
			}
			break OpLoop
		default:
			return nil, fmt.Errorf("Unknown create status %q: %+v", op.Status, op)
		}
	}
	condRun(opts.OnInstanceCreated)

	apiGate()
	inst, err := computeService.Instances.Get(projectID, zone, instName).Do()
	if err != nil {
		return nil, fmt.Errorf("Error getting instance %s details after creation: %v", instName, err)
	}

	// Finds its internal and/or external IP addresses.
	intIP, extIP := instanceIPs(inst)

	// Wait for it to boot and its buildlet to come up.
	var buildletURL string
	var ipPort string
	if !opts.TLS.IsZero() {
		if extIP == "" {
			return nil, errors.New("didn't find its external IP address")
		}
		buildletURL = "https://" + extIP
		ipPort = extIP + ":443"
	} else {
		if intIP == "" {
			return nil, errors.New("didn't find its internal IP address")
		}
		buildletURL = "http://" + intIP
		ipPort = intIP + ":80"
	}
	condRun(opts.OnGotInstanceInfo)

	const timeout = 3 * time.Minute
	var alive bool
	impatientClient := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			Dial:              defaultDialer(),
			DisableKeepAlives: true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	deadline := time.Now().Add(timeout)
	try := 0
	for time.Now().Before(deadline) {
		try++
		res, err := impatientClient.Get(buildletURL)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		res.Body.Close()
		if res.StatusCode != 200 {
			return nil, fmt.Errorf("buildlet returned HTTP status code %d on try number %d", res.StatusCode, try)
		}
		alive = true
		break
	}
	if !alive {
		return nil, fmt.Errorf("buildlet didn't come up at %s in %v", buildletURL, timeout)
	}

	return NewClient(ipPort, opts.TLS), nil
}

// DestroyVM sends a request to delete a VM. Actual VM description is
// currently (2015-01-19) very slow for no good reason. This function
// returns once it's been requested, not when it's done.
func DestroyVM(ts oauth2.TokenSource, proj, zone, instance string) error {
	computeService, _ := compute.New(oauth2.NewClient(oauth2.NoContext, ts))
	apiGate()
	_, err := computeService.Instances.Delete(proj, zone, instance).Do()
	return err
}

type VM struct {
	// Name is the name of the GCE VM instance.
	// For example, it's of the form "mote-bradfitz-plan9-386-foo",
	// and not "plan9-386-foo".
	Name   string
	IPPort string
	TLS    KeyPair
	Type   string
}

// ListVMs lists all VMs.
func ListVMs(ts oauth2.TokenSource, proj, zone string) ([]VM, error) {
	var vms []VM
	computeService, _ := compute.New(oauth2.NewClient(oauth2.NoContext, ts))

	// TODO(bradfitz): paging over results if more than 500
	apiGate()
	list, err := computeService.Instances.List(proj, zone).Do()
	if err != nil {
		return nil, err
	}
	for _, inst := range list.Items {
		if inst.Metadata == nil {
			// Defensive. Not seen in practice.
			continue
		}
		meta := map[string]string{}
		for _, it := range inst.Metadata.Items {
			if it.Value != nil {
				meta[it.Key] = *it.Value
			}
		}
		builderType := meta["builder-type"]
		if builderType == "" {
			continue
		}
		vm := VM{
			Name: inst.Name,
			Type: builderType,
			TLS: KeyPair{
				CertPEM: meta["tls-cert"],
				KeyPEM:  meta["tls-key"],
			},
		}
		_, extIP := instanceIPs(inst)
		if extIP == "" || vm.TLS.IsZero() {
			continue
		}
		vm.IPPort = extIP + ":443"
		vms = append(vms, vm)
	}
	return vms, nil
}

func instanceIPs(inst *compute.Instance) (intIP, extIP string) {
	for _, iface := range inst.NetworkInterfaces {
		if strings.HasPrefix(iface.NetworkIP, "10.") {
			intIP = iface.NetworkIP
		}
		for _, accessConfig := range iface.AccessConfigs {
			if accessConfig.Type == "ONE_TO_ONE_NAT" {
				extIP = accessConfig.NatIP
			}
		}
	}
	return
}
