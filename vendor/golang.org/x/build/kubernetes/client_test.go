package kubernetes_test

import (
	"log"
	"net/http"

	api "github.com/kubernetes/kubernetes/pkg/api/v1"
	"golang.org/x/build/kubernetes"
	"golang.org/x/oauth2"
)

func ExampleRun() {
	kube, err := kubernetes.NewClient("example.com", &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "aCcessWbU3toKen"}),
		}})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	kube.Run(&api.Pod{
		TypeMeta: api.TypeMeta{
			APIVersion: "v1beta3",
			Kind:       "Pod",
		},
		ObjectMeta: api.ObjectMeta{
			Name: "my-nginx-pod",
			Labels: map[string]string{
				"tag": "prod",
			},
		},
		Spec: api.PodSpec{
			Containers: []api.Container{
				{
					Name:  "my-nginx-container",
					Image: "nginx:latest",
				},
			},
		},
	})
}
