// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main // import "golang.org/x/build/cmd/coordinator/buildongce"

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"go4.org/cloud/google/gceutil"

	"golang.org/x/build/buildenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	dm "google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/googleapi"
	"google.golang.org/cloud/datastore"
)

var (
	proj            = flag.String("project", "", "Optional name of the Google Cloud Platform project to create the infrastructure in. If empty, the project defined in golang.org/x/build/buildenv is used, for either production or staging (if the -staging flag is used)")
	staticIP        = flag.String("static_ip", "", "Static IP to use. If empty, automatic.")
	reuseDisk       = flag.Bool("reuse_disk", true, "Whether disk images should be reused between shutdowns/restarts.")
	ssd             = flag.Bool("ssd", true, "If true, use a solid state disk (faster, more expensive)")
	coordinator     = flag.String("coord", "", "Optional coordinator binary URL. If empty, the URL from a configuration defined in golang.org/x/build/buildenv will be used. ")
	staging         = flag.Bool("staging", false, "If true, buildenv.Staging will be used to provide default configuration values. Otherwise, buildenv.Production is used.")
	skipKube        = flag.Bool("skip_kube", false, "If true, the Kubernetes cluster will not be created.")
	skipCoordinator = flag.Bool("skip_coordinator", false, "If true, the coordinator instance will not be created.")
	makeDisks       = flag.Bool("make_basepin", false, "Create the basepin disk images for all builders, then stop. Does not create the VM.")

	computeService    *compute.Service
	deploymentService *dm.Service
	oauthClient       *http.Client
	err               error
	buildEnv          *buildenv.Environment
)

const baseConfig = `#cloud-config
coreos:
  update:
    group: stable
    reboot-strategy: off
  units:
    - name: gobuild.service
      command: start
      content: |
        [Unit]
        Description=Go Builders
        After=docker.service
        Requires=docker.service
        
        [Service]
        ExecStartPre=/bin/bash -c 'mkdir -p /opt/bin && curl -s -o /opt/bin/coordinator.tmp $COORDINATOR && install -m 0755 /opt/bin/coordinator{.tmp,}'
        ExecStart=/opt/bin/coordinator
        RestartSec=10s
        Restart=always
        StartLimitInterval=0
        Type=simple
         
        [Install]
        WantedBy=multi-user.target
`

// Deployment Manager V2 manifest for creating a Google Container Engine
// cluster to run buildlets, as well as an autoscaler attached to the
// cluster's instance group to add capacity based on CPU utilization
const kubeConfig = `
resources:
- name: "{{ .KubeName }}"
  type: container.v1.cluster
  properties:
    zone: "{{ .Zone }}"
    cluster:
      initial_node_count: {{ .KubeMinNodes }}
      network: "default"
      logging_service: "logging.googleapis.com"
      monitoring_service: "none"
      node_config:
        machine_type: "{{ .KubeMachineType }}"
        oauth_scopes:
          - "https://www.googleapis.com/auth/cloud-platform"
      master_auth:
        username: "admin"
        password: "{{ .KubePassword }}"
- name: autoscaler
  type: compute.v1.autoscaler
  properties:
    zone: "{{ .Zone }}"
    name: "{{ .KubeName }}"
    target: "$(ref.{{ .KubeName }}.instanceGroupUrls[0])"
    autoscalingPolicy:
      minNumReplicas: {{ .KubeMinNodes }}
      maxNumReplicas: {{ .KubeMaxNodes }}
      coolDownPeriodSec: 1200
      cpuUtilization:
        utilizationTarget: .6`

func readFile(v string) string {
	slurp, err := ioutil.ReadFile(v)
	if err != nil {
		log.Fatalf("Error reading %s: %v", v, err)
	}
	return strings.TrimSpace(string(slurp))
}

func main() {
	buildEnv = buildenv.Production

	flag.Parse()

	if *staging {
		buildEnv = buildenv.Staging
	}
	if *proj != "" {
		buildEnv.ProjectName = *proj
	}
	if *coordinator != "" {
		buildEnv.CoordinatorURL = *coordinator
	}
	if *staticIP != "" {
		buildEnv.StaticIP = *staticIP
	}

	buildEnv.KubePassword = randomPassword()

	// Brad is sick of google.DefaultClient giving him the
	// permissions from the instance via the metadata service. Use
	// the service account from disk if it exists instead:
	keyFile := filepath.Join(os.Getenv("HOME"), "keys", buildEnv.ProjectName+".key.json")
	if _, err := os.Stat(keyFile); err == nil {
		log.Printf("Using service account from %s", keyFile)
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", keyFile)
	}

	oauthClient, err = google.DefaultClient(oauth2.NoContext, compute.CloudPlatformScope, compute.ComputeScope, compute.DevstorageFullControlScope)
	if err != nil {
		log.Fatalf("could not create oAuth client: %v", err)
	}

	computeService, err = compute.New(oauthClient)
	if err != nil {
		log.Fatalf("could not create client for Google Compute Engine: %v", err)
	}

	if *makeDisks {
		if err := makeBasepinDisks(computeService); err != nil {
			log.Fatalf("could not create basepin disks: %v", err)
		}
		return
	}

	if !*skipCoordinator {
		err = createCoordinator()
		if err != nil {
			log.Fatalf("Error creating coordinator instance: %v", err)
		}
	}

	if !*skipKube {
		err = createCluster()
		if err != nil {
			log.Fatalf("Error creating Kubernetes cluster: %v", err)
		}
	}
}

func createCoordinator() error {
	log.Printf("Creating coordinator instance: %v", buildEnv.CoordinatorName)

	natIP := buildEnv.StaticIP
	if natIP == "" {
		// Try to find it by name.
		aggAddrList, err := computeService.Addresses.AggregatedList(buildEnv.ProjectName).Do()
		if err != nil {
			return fmt.Errorf("could not find IP address: %v", err)
		}
		// https://godoc.org/google.golang.org/api/compute/v1#AddressAggregatedList
	IPLoop:
		for _, asl := range aggAddrList.Items {
			for _, addr := range asl.Addresses {
				if addr.Name == buildEnv.CoordinatorName+"-ip" && addr.Status == "RESERVED" {
					natIP = addr.Address
					break IPLoop
				}
			}
		}
	}

	cloudConfig := strings.Replace(baseConfig, "$COORDINATOR", buildEnv.CoordinatorURL, 1)
	const maxCloudConfig = 32 << 10 // per compute API docs
	if len(cloudConfig) > maxCloudConfig {
		return fmt.Errorf("cloud config length of %d bytes is over %d byte limit", len(cloudConfig), maxCloudConfig)
	}

	instance := &compute.Instance{
		Name:        buildEnv.CoordinatorName,
		Description: "Go Builder",
		MachineType: buildEnv.MachineTypeURI(),
		Disks:       []*compute.AttachedDisk{instanceDisk(computeService)},
		Tags: &compute.Tags{
			Items: []string{"http-server", "https-server", "allow-ssh"},
		},
		Metadata: &compute.Metadata{
			Items: []*compute.MetadataItems{
				{
					Key:   "user-data",
					Value: googleapi.String(cloudConfig),
				},
			},
		},
		NetworkInterfaces: []*compute.NetworkInterface{
			&compute.NetworkInterface{
				AccessConfigs: []*compute.AccessConfig{
					{
						Type:  "ONE_TO_ONE_NAT",
						Name:  "External NAT",
						NatIP: natIP,
					},
				},
				Network: buildEnv.ComputePrefix() + "/global/networks/default",
			},
		},
		ServiceAccounts: []*compute.ServiceAccount{
			{
				Email: "default",
				Scopes: []string{
					compute.DevstorageFullControlScope,
					compute.ComputeScope,
					compute.CloudPlatformScope,
					datastore.ScopeDatastore,
				},
			},
		},
	}

	log.Printf("Creating instance...")
	op, err := computeService.Instances.Insert(buildEnv.ProjectName, buildEnv.Zone, instance).Do()
	if err != nil {
		return fmt.Errorf("Failed to create instance: %v", err)
	}
	if err := awaitOp(computeService, op); err != nil {
		log.Fatalf("failed to start: %v", err)
	}

	inst, err := computeService.Instances.Get(buildEnv.ProjectName, buildEnv.Zone, buildEnv.CoordinatorName).Do()
	if err != nil {
		log.Fatalf("Error getting instance after creation: %v", err)
	}
	ij, _ := json.MarshalIndent(inst, "", "    ")
	log.Printf("Instance: %s", ij)
	return nil
}

func awaitOp(svc *compute.Service, op *compute.Operation) error {
	opName := op.Name
	log.Printf("Waiting on operation %v", opName)
	for {
		time.Sleep(2 * time.Second)
		op, err := svc.ZoneOperations.Get(buildEnv.ProjectName, buildEnv.Zone, opName).Do()
		if err != nil {
			return fmt.Errorf("Failed to get op %s: %v", opName, err)
		}
		switch op.Status {
		case "PENDING", "RUNNING":
			log.Printf("Waiting on operation %v", opName)
			continue
		case "DONE":
			if op.Error != nil {
				var last error
				for _, operr := range op.Error.Errors {
					log.Printf("Error: %+v", operr)
					last = fmt.Errorf("%v", operr)
				}
				return last
			}
			log.Printf("Success. %+v", op)
			return nil
		default:
			return fmt.Errorf("Unknown status %q: %+v", op.Status, op)
		}
	}
}

func createCluster() error {
	log.Printf("Creating Kubernetes cluster: %v", buildEnv.KubeName)
	deploymentService, err = dm.New(oauthClient)
	if err != nil {
		return fmt.Errorf("could not create client for Google Cloud Deployment Manager: %v", err)
	}

	if buildEnv.KubeMaxNodes == 0 || buildEnv.KubeMinNodes == 0 {
		return fmt.Errorf("buildenv KubeMaxNodes/KubeMinNodes values cannot be 0")
	}

	tpl, err := template.New("kube").Parse(kubeConfig)
	if err != nil {
		return fmt.Errorf("could not parse Deployment Manager template: %v", err)
	}

	var result bytes.Buffer
	err = tpl.Execute(&result, buildEnv)
	if err != nil {
		return fmt.Errorf("could not execute Deployment Manager template: %v", err)
	}

	deployment := &dm.Deployment{
		Name: buildEnv.KubeName,
		Target: &dm.TargetConfiguration{
			Config: &dm.ConfigFile{
				Content: result.String(),
			},
		},
	}
	op, err := deploymentService.Deployments.Insert(buildEnv.ProjectName, deployment).Do()
	if err != nil {
		return fmt.Errorf("Failed to create cluster with Deployment Manager: %v", err)
	}
	opName := op.Name
	log.Printf("Created. Waiting on operation %v", opName)
OpLoop:
	for {
		time.Sleep(2 * time.Second)
		op, err := deploymentService.Operations.Get(buildEnv.ProjectName, opName).Do()
		if err != nil {
			return fmt.Errorf("Failed to get op %s: %v", opName, err)
		}
		switch op.Status {
		case "PENDING", "RUNNING":
			log.Printf("Waiting on operation %v", opName)
			continue
		case "DONE":
			// If no errors occurred, op.StatusMessage is empty.
			if op.StatusMessage != "" {
				log.Printf("Error: %+v", op.StatusMessage)
				return fmt.Errorf("Failed to create.")
			}
			log.Printf("Success.")
			break OpLoop
		default:
			return fmt.Errorf("Unknown status %q: %+v", op.Status, op)
		}
	}
	return nil
}

func instanceDisk(svc *compute.Service) *compute.AttachedDisk {
	imageURL, err := gceutil.CoreOSImageURL(oauthClient)
	if err != nil {
		log.Fatalf("Error fetching CoreOS Image URL: %v", err)
	}
	diskName := buildEnv.CoordinatorName + "-coreos-stateless-pd"

	if *reuseDisk {
		dl, err := svc.Disks.List(buildEnv.ProjectName, buildEnv.Zone).Do()
		if err != nil {
			log.Fatalf("Error listing disks: %v", err)
		}
		for _, disk := range dl.Items {
			if disk.Name != diskName {
				continue
			}
			return &compute.AttachedDisk{
				AutoDelete: false,
				Boot:       true,
				DeviceName: diskName,
				Type:       "PERSISTENT",
				Source:     disk.SelfLink,
				Mode:       "READ_WRITE",

				// The GCP web UI's "Show REST API" link includes a
				// "zone" parameter, but it's not in the API
				// description. But it wants this form (disk.Zone, a
				// full zone URL, not *zone):
				// Zone: disk.Zone,
				// ... but it seems to work without it.  Keep this
				// comment here until I file a bug with the GCP
				// people.
			}
		}
	}

	diskType := ""
	if *ssd {
		diskType = "https://www.googleapis.com/compute/v1/projects/" + buildEnv.ProjectName + "/zones/" + buildEnv.Zone + "/diskTypes/pd-ssd"
	}

	return &compute.AttachedDisk{
		AutoDelete: !*reuseDisk,
		Boot:       true,
		Type:       "PERSISTENT",
		InitializeParams: &compute.AttachedDiskInitializeParams{
			DiskName:    diskName,
			SourceImage: imageURL,
			DiskSizeGb:  50,
			DiskType:    diskType,
		},
	}
}

func randomPassword() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	p := make([]byte, 20)
	rand.Seed(time.Now().UnixNano())
	for i := range p {
		p[i] = chars[rand.Intn(len(chars))]
	}
	return string(p)
}

func makeBasepinDisks(svc *compute.Service) error {
	// Try to find it by name.
	imList, err := svc.Images.List(buildEnv.ProjectName).Do()
	if err != nil {
		return fmt.Errorf("Error listing images for %s: %v", buildEnv.ProjectName, err)
	}
	if imList.NextPageToken != "" {
		return errors.New("too many images; pagination not supported")
	}
	diskList, err := svc.Disks.List(buildEnv.ProjectName, buildEnv.Zone).Do()
	if err != nil {
		return err
	}
	if diskList.NextPageToken != "" {
		return errors.New("too many disks; pagination not supported (yet?)")
	}

	need := make(map[string]*compute.Image) // keys like "https://www.googleapis.com/compute/v1/projects/symbolic-datum-552/global/images/linux-buildlet-arm"
	for _, im := range imList.Items {
		need[im.SelfLink] = im
	}

	for _, d := range diskList.Items {
		if !strings.HasPrefix(d.Name, "basepin-") {
			continue
		}
		if si, ok := need[d.SourceImage]; ok && d.SourceImageId == fmt.Sprint(si.Id) {
			log.Printf("Have %s: %s (%v)\n", d.Name, d.SourceImage, d.SourceImageId)
			delete(need, d.SourceImage)
		}
	}

	var needed []string
	for imageName := range need {
		needed = append(needed, imageName)
	}
	sort.Strings(needed)
	for _, n := range needed {
		log.Printf("Need %v", n)
	}
	for i, imName := range needed {
		im := need[imName]
		log.Printf("(%d/%d) Creating %s ...", i+1, len(needed), im.Name)
		op, err := svc.Disks.Insert(buildEnv.ProjectName, buildEnv.Zone, &compute.Disk{
			Description:   "zone-cached basepin image of " + im.Name,
			Name:          "basepin-" + im.Name + "-" + fmt.Sprint(im.Id),
			SizeGb:        im.DiskSizeGb,
			SourceImage:   im.SelfLink,
			SourceImageId: fmt.Sprint(im.Id),
			Type:          "https://www.googleapis.com/compute/v1/projects/" + buildEnv.ProjectName + "/zones/" + buildEnv.Zone + "/diskTypes/pd-ssd",
		}).Do()
		if err != nil {
			return err
		}
		if err := awaitOp(svc, op); err != nil {
			log.Fatalf("failed to create: %v", err)
		}
	}
	return nil
}
