package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var (
	apiHost              = "http://127.0.0.1:8001"
	apiPrefix            = "v1"
	resourceKind         = "pods"
	caCertificate        = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	bearerTokenPath      = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	usePodServiceAccount = false
)

type ResourceEvent struct {
	Type   string   `json:"type"`
	Object Resource `json:"object"`
}

type ResourceList struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   map[string]interface{} `json:"metadata"`
	Items      []Resource             `json:"items"`
}

type Resource struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   ResourceMetadata       `json:"metadata"`
	Spec       map[string]interface{} `json:"spec"`
}

type ResourceMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	SelfLink  string `json:"selfLink"`
}

func httpClient() (*http.Client, error) {
	certs := x509.NewCertPool()
	pemData, err := ioutil.ReadFile(caCertificate)
	if err != nil {
		return nil, err
	}
	certs.AppendCertsFromPEM(pemData)

	newTLSConfig := &tls.Config{}
	newTLSConfig.RootCAs = certs

	tr := &http.Transport{TLSClientConfig: newTLSConfig}
	client := &http.Client{Transport: tr}
	return client, err
}

func httpGet(url string) (*http.Response, error) {
	client, err := httpClient()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(bearerTokenPath)
	if err != nil {
		return nil, err
	}

	token, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+string(token))
	return client.Do(req)
}

func url() string {
	if apiPrefix != "v1" {
		return apiHost + "/apis/" + apiPrefix + "/" + resourceKind
	}
	return apiHost + "/api/v1/" + resourceKind
}

func watchurl() string {
	return url() + "?watch=true"
}

func getResources() ([]Resource, error) {
	resp, err := httpGet(url())
	if err != nil {
		return nil, err
	}

	var resourceList ResourceList
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&resourceList)
	if err != nil {
		return nil, err
	}

	return resourceList.Items, nil
}

func watchResourceEvents() (<-chan ResourceEvent, <-chan error) {
	events := make(chan ResourceEvent)
	errs := make(chan error, 1)

	go func() {
		for {
			resp, err := httpGet(watchurl())
			if err != nil {
				errs <- err
				time.Sleep(5 * time.Second)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				errs <- errors.New("invalid status code: " + resp.Status)
				time.Sleep(5 * time.Second)
				continue
			}

			decoder := json.NewDecoder(resp.Body)
			for {
				var event ResourceEvent
				err = decoder.Decode(&event)
				if err != nil {
					errs <- err
					break
				}
				events <- event
			}
		}
	}()

	return events, errs
}
