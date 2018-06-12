// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package buildenv contains definitions for the
// environments the Go build system can run in.
package buildenv

import (
	"fmt"
	"strings"
)

const (
	prefix = "https://www.googleapis.com/compute/v1/projects/"
)

// Environment describes the configuration of the infrastructure for a
// coordinator and its buildlet resources running on Google Cloud Platform.
// Staging and Production are the two common build environments.
type Environment struct {
	// The GCP project name that the build infrastructure will be provisioned in.
	// This field may be overridden as necessary without impacting other fields.
	ProjectName string

	// The IsProd flag indicates whether production functionality should be
	// enabled. When true, GCE and Kubernetes builders are enabled and the
	// coordinator serves on 443. Otherwise, GCE and Kubernetes builders are
	// disabled and the coordinator serves on 8119.
	IsProd bool

	// Zone is the GCE zone that the coordinator instance and Kubernetes cluster
	// will run in. This field may be overridden as necessary without impacting
	// other fields.
	Zone string

	// ZonesToClean are the GCE zones that will be periodically cleaned by
	// deleting old VMs. The zero value means that no cleaning will occur.
	// This field is optional.
	ZonesToClean []string

	// StaticIP is the public, static IP address that will be attached to the
	// coordinator instance. The zero value means the address will be looked
	// up by name. This field is optional.
	StaticIP string

	// MachineType is the GCE machine type to use for the coordinator.
	MachineType string

	// KubeMinNodes is the minimum number of nodes in the Kubernetes cluster.
	// The autoscaler will ensure that at least this many nodes is always
	// running despite any scale-down decision.
	KubeMinNodes int64

	// KubeMaxNodes is the maximum number of nodes that the autoscaler can
	// provision in the Kubernetes cluster.
	// If KubeMaxNodes is 0, Kubernetes is not used.
	KubeMaxNodes int64

	// KubeMachineType is the GCE machine type to use for the Kubernetes cluster nodes.
	KubeMachineType string

	// KubeName is the name of the Kubernetes cluster that will be created.
	KubeName string

	// KubePassword is the admin password for the Kubernetes cluster. Its value
	// will be set to a random value at runtime.
	KubePassword string

	// CoordinatorURL is the location from which the coordinator
	// binary will be downloaded.
	// This is only used by cmd/coordinator/buildongce/create.go when
	// creating the coordinator VM from scratch.
	CoordinatorURL string

	// CoordinatorName is the hostname of the coordinator instance.
	CoordinatorName string

	// BuildletBucket is the GCS bucket that stores buildlet binaries.
	// TODO: rename. this is not just for buildlets; also for bootstrap.
	BuildletBucket string

	// LogBucket is the GCS bucket to which logs are written.
	LogBucket string

	// SnapBucket is the GCS bucket to which snapshots of
	// completed builds (after make.bash, before tests) are
	// written.
	SnapBucket string

	// MaxBuilds is the maximum number of concurrent builds that
	// can run. Zero means unlimit. This is typically only used
	// in a development or staging environment.
	MaxBuilds int
}

// MachineTypeURI returns the URI for the environment's Machine Type.
func (e Environment) MachineTypeURI() string {
	return e.ComputePrefix() + "/zones/" + e.Zone + "/machineTypes/" + e.MachineType
}

// ComputePrefix returns the URI prefix for Compute Engine resources in a project.
func (e Environment) ComputePrefix() string {
	return prefix + e.ProjectName
}

// Region returns the GCE region, derived from its zone.
func (e Environment) Region() string {
	return e.Zone[:strings.LastIndex(e.Zone, "-")]
}

// ByProjectID returns an Environment for the specified
// project ID. It is currently limited to the symbolic-datum-552
// and go-dashboard-dev projects.
// ByProjectID will panic if the project ID is not known.
func ByProjectID(projectID string) *Environment {
	var envKeys []string

	for k := range possibleEnvs {
		envKeys = append(envKeys, k)
	}

	var env *Environment
	env, ok := possibleEnvs[projectID]
	if !ok {
		panic(fmt.Sprintf("Can't get buildenv for unknown project %q. Possible envs are %s", projectID, envKeys))
	}

	return env
}

// Staging defines the environment that the coordinator and build
// infrastructure is deployed to before it is released to production.
// For local dev, override the project with the program's flag to set
// a custom project.
var Staging = &Environment{
	ProjectName:     "go-dashboard-dev",
	IsProd:          true,
	Zone:            "us-central1-f",
	ZonesToClean:    []string{"us-central1-a", "us-central1-b", "us-central1-f"},
	StaticIP:        "104.154.113.235",
	MachineType:     "n1-standard-1",
	KubeMinNodes:    1,
	KubeMaxNodes:    2,
	KubeName:        "buildlets",
	KubeMachineType: "n1-standard-32",
	CoordinatorURL:  "https://storage.googleapis.com/dev-go-builder-data/coordinator",
	CoordinatorName: "farmer",
	BuildletBucket:  "dev-go-builder-data",
	LogBucket:       "dev-go-build-log",
	SnapBucket:      "dev-go-build-snap",
}

// Production defines the environment that the coordinator and build
// infrastructure is deployed to for production usage at build.golang.org.
var Production = &Environment{
	ProjectName:     "symbolic-datum-552",
	IsProd:          true,
	Zone:            "us-central1-f",
	ZonesToClean:    []string{"us-central1-f"},
	StaticIP:        "107.178.219.46",
	MachineType:     "n1-standard-4",
	KubeMinNodes:    3,
	KubeMaxNodes:    5,
	KubeName:        "buildlets",
	KubeMachineType: "n1-standard-32",
	CoordinatorURL:  "https://storage.googleapis.com/go-builder-data/coordinator",
	CoordinatorName: "farmer",
	BuildletBucket:  "go-builder-data",
	LogBucket:       "go-build-log",
	SnapBucket:      "go-build-snap",
}

var Development = &Environment{
	IsProd:   false,
	StaticIP: "127.0.0.1",
}

// possibleEnvs enumerate the known buildenv.Environment definitions.
var possibleEnvs = map[string]*Environment{
	"dev":                Development,
	"symbolic-datum-552": Production,
	"go-dashboard-dev":   Staging,
}
