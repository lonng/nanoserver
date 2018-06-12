// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package kubernetes contains a minimal client for the Kubernetes API.
package kubernetes

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/build/kubernetes/api"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

const (
	// APIEndpoint defines the base path for kubernetes API resources.
	APIEndpoint     = "/api/v1"
	defaultPod      = "/namespaces/default/pods"
	defaultWatchPod = "/watch/namespaces/default/pods"
	nodes           = "/nodes"
)

// Client is a client for the Kubernetes master.
type Client struct {
	endpointURL string
	httpClient  *http.Client
}

// NewClient returns a new Kubernetes client.
// The provided host is an url (scheme://hostname[:port]) of a
// Kubernetes master without any path.
// The provided client is an authorized http.Client used to perform requests to the Kubernetes API master.
func NewClient(baseURL string, client *http.Client) (*Client, error) {
	validURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL %q: %v", baseURL, err)
	}
	return &Client{
		endpointURL: strings.TrimSuffix(validURL.String(), "/") + APIEndpoint,
		httpClient:  client,
	}, nil
}

// RunLongLivedPod creates a new pod resource in the default pod namespace with
// the given pod API specification. It assumes the pod runs a
// long-lived server (i.e. if the container exit quickly quickly, even
// with success, then that is an error).
//
// It returns the pod status once it has entered the Running phase.
// An error is returned if the pod can not be created, or if ctx.Done
// is closed.
func (c *Client) RunLongLivedPod(ctx context.Context, pod *api.Pod) (*api.PodStatus, error) {
	var podJSON bytes.Buffer
	if err := json.NewEncoder(&podJSON).Encode(pod); err != nil {
		return nil, fmt.Errorf("failed to encode pod in json: %v", err)
	}
	postURL := c.endpointURL + defaultPod
	req, err := http.NewRequest("POST", postURL, &podJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: POST %q : %v", postURL, err)
	}
	res, err := ctxhttp.Do(ctx, c.httpClient, req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: POST %q: %v", postURL, err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read request body for POST %q: %v", postURL, err)
	}
	if res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("http error: %d POST %q: %q: %v", res.StatusCode, postURL, string(body), err)
	}
	var podResult api.Pod
	if err := json.Unmarshal(body, &podResult); err != nil {
		return nil, fmt.Errorf("failed to decode pod resources: %v", err)
	}

	for {
		// TODO(bradfitz,evanbrown): pass podResult.ObjectMeta.ResourceVersion to PodStatus?
		ps, err := c.PodStatus(ctx, podResult.Name)
		if err != nil {
			return nil, err
		}
		switch ps.Phase {
		case api.PodPending:
			// The main phase we're waiting on
			break
		case api.PodRunning:
			return ps, nil
		case api.PodSucceeded, api.PodFailed:
			return nil, fmt.Errorf("pod entered phase %q", ps.Phase)
		default:
			log.Printf("RunLongLivedPod poll loop: pod %q in unexpected phase %q; sleeping", podResult.Name, ps.Phase)
		}
		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			// The pod did not leave the pending
			// state. Try to clean it up.
			go c.DeletePod(context.Background(), podResult.Name)
			return nil, ctx.Err()
		}
	}
}

// GetPods returns all pods in the cluster, regardless of status.
func (c *Client) GetPods(ctx context.Context) ([]api.Pod, error) {
	getURL := c.endpointURL + defaultPod

	// Make request to Kubernetes API
	req, err := http.NewRequest("GET", getURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: GET %q : %v", getURL, err)
	}
	res, err := ctxhttp.Do(ctx, c.httpClient, req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: GET %q: %v", getURL, err)
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read request body for GET %q: %v", getURL, err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error %d GET %q: %q: %v", res.StatusCode, getURL, string(body), err)
	}

	var podList api.PodList
	if err := json.Unmarshal(body, &podList); err != nil {
		return nil, fmt.Errorf("failed to decode list of pod resources: %v", err)
	}
	return podList.Items, nil
}

// PodDelete deletes the specified Kubernetes pod.
func (c *Client) DeletePod(ctx context.Context, podName string) error {
	url := c.endpointURL + defaultPod + "/" + podName
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: DELETE %q : %v", url, err)
	}
	res, err := ctxhttp.Do(ctx, c.httpClient, req)
	if err != nil {
		return fmt.Errorf("failed to make request: DELETE %q: %v", url, err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read response body: DELETE %q: %v, url, err")
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("http error: %d DELETE %q: %q: %v", res.StatusCode, url, string(body), err)
	}
	return nil
}

// TODO(bradfitz): WatchPod is unreliable, so this is disabled.
//
// AwaitPodNotPending will return a pod's status in a
// podStatusResult when the pod is no longer in the pending
// state.
// The podResourceVersion is required to prevent a pod's entire
// history from being retrieved when the watch is initiated.
// If there is an error polling for the pod's status, or if
// ctx.Done is closed, podStatusResult will contain an error.
func (c *Client) _AwaitPodNotPending(ctx context.Context, podName, podResourceVersion string) (*api.Pod, error) {
	if podResourceVersion == "" {
		return nil, fmt.Errorf("resourceVersion for pod %v must be provided", podName)
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	podStatusUpdates, err := c._WatchPod(ctx, podName, podResourceVersion)
	if err != nil {
		return nil, err
	}
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case psr := <-podStatusUpdates:
			if psr.Err != nil {
				// If the context is done, prefer its error:
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
					return nil, psr.Err
				}
			}
			if psr.Pod.Status.Phase != api.PodPending {
				return psr.Pod, nil
			}
		}
	}
}

// PodStatusResult wraps an api.PodStatus and error.
type PodStatusResult struct {
	Pod  *api.Pod
	Type string
	Err  error
}

type watchPodStatus struct {
	// The type of watch update contained in the message
	Type string `json:"type"`
	// Pod details
	Object api.Pod `json:"object"`
}

// TODO(bradfitz): WatchPod is unreliable and sometimes hangs forever
// without closing and sometimes ends prematurely, so this API is
// disabled.
//
// WatchPod long-polls the Kubernetes watch API to be notified
// of changes to the specified pod. Changes are sent on the returned
// PodStatusResult channel as they are received.
// The podResourceVersion is required to prevent a pod's entire
// history from being retrieved when the watch is initiated.
// The provided context must be canceled or timed out to stop the watch.
// If any error occurs communicating with the Kubernetes API, the
// error will be sent on the returned PodStatusResult channel and
// it will be closed.
func (c *Client) _WatchPod(ctx context.Context, podName, podResourceVersion string) (<-chan PodStatusResult, error) {
	if podResourceVersion == "" {
		return nil, fmt.Errorf("resourceVersion for pod %v must be provided", podName)
	}
	statusChan := make(chan PodStatusResult, 1)

	go func() {
		defer close(statusChan)
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		// Make request to Kubernetes API
		getURL := c.endpointURL + defaultWatchPod + "/" + podName
		req, err := http.NewRequest("GET", getURL, nil)
		req.URL.Query().Add("resourceVersion", podResourceVersion)
		if err != nil {
			statusChan <- PodStatusResult{Err: fmt.Errorf("failed to create request: GET %q : %v", getURL, err)}
			return
		}
		res, err := ctxhttp.Do(ctx, c.httpClient, req)
		if err != nil {
			statusChan <- PodStatusResult{Err: err}
			return
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			statusChan <- PodStatusResult{Err: fmt.Errorf("WatchPod status %v", res.Status)}
			return
		}
		reader := bufio.NewReader(res.Body)

		// bufio.Reader.ReadBytes is blocking, so we watch for
		// context timeout or cancellation in a goroutine
		// and close the response body when see see it. The
		// response body is also closed via defer when the
		// request is made, but closing twice is OK.
		go func() {
			<-ctx.Done()
			res.Body.Close()
		}()

		const backupPollDuration = 30 * time.Second
		backupPoller := time.AfterFunc(backupPollDuration, func() {
			log.Printf("kubernetes: backup poller in WatchPod checking on %q", podName)
			st, err := c.PodStatus(ctx, podName)
			log.Printf("kubernetes: backup poller in WatchPod PodStatus(%q) = %v, %v", podName, st, err)
			if err != nil {
				// Some error.
				cancel()
			}
		})
		defer backupPoller.Stop()

		for {
			line, err := reader.ReadBytes('\n')
			log.Printf("kubernetes WatchPod status line of %q: %q, %v", podName, line, err)
			backupPoller.Reset(backupPollDuration)
			if err != nil {
				statusChan <- PodStatusResult{Err: fmt.Errorf("error reading streaming response body: %v", err)}
				return
			}
			var wps watchPodStatus
			if err := json.Unmarshal(line, &wps); err != nil {
				statusChan <- PodStatusResult{Err: fmt.Errorf("failed to decode watch pod status: %v", err)}
				return
			}
			statusChan <- PodStatusResult{Pod: &wps.Object, Type: wps.Type}
		}
	}()
	return statusChan, nil
}

// Retrieve the status of a pod synchronously from the Kube
// API server.
func (c *Client) PodStatus(ctx context.Context, podName string) (*api.PodStatus, error) {
	getURL := c.endpointURL + defaultPod + "/" + podName

	// Make request to Kubernetes API
	req, err := http.NewRequest("GET", getURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: GET %q : %v", getURL, err)
	}
	res, err := ctxhttp.Do(ctx, c.httpClient, req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: GET %q: %v", getURL, err)
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read request body for GET %q: %v", getURL, err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error %d GET %q: %q: %v", res.StatusCode, getURL, string(body), err)
	}

	var pod *api.Pod
	if err := json.Unmarshal(body, &pod); err != nil {
		return nil, fmt.Errorf("failed to decode pod resources: %v", err)
	}
	return &pod.Status, nil
}

// PodLog retrieves the container log for the first container
// in the pod.
func (c *Client) PodLog(ctx context.Context, podName string) (string, error) {
	// TODO(evanbrown): support multiple containers
	url := c.endpointURL + defaultPod + "/" + podName + "/log"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: GET %q : %v", url, err)
	}
	res, err := ctxhttp.Do(ctx, c.httpClient, req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: GET %q: %v", url, err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", fmt.Errorf("failed to read response body: GET %q: %v", url, err)
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http error %d GET %q: %q: %v", res.StatusCode, url, string(body), err)
	}
	return string(body), nil
}

// PodNodes returns the list of nodes that comprise the Kubernetes cluster
func (c *Client) GetNodes(ctx context.Context) ([]api.Node, error) {
	url := c.endpointURL + nodes
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: GET %q : %v", url, err)
	}
	res, err := ctxhttp.Do(ctx, c.httpClient, req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: GET %q: %v", url, err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: GET %q: %v, url, err")
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error %d GET %q: %q: %v", res.StatusCode, url, string(body), err)
	}
	var nodeList *api.NodeList
	if err := json.Unmarshal(body, &nodeList); err != nil {
		return nil, fmt.Errorf("failed to decode node list: %v", err)
	}
	return nodeList.Items, nil
}
