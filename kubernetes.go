package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type NamespaceList struct {
	Items []Namespace `json:"items"`
}

type Namespace struct {
	Metadata Metadata `json:"metadata"`
}

type ConfigMapList struct {
	Items []ConfigMap `json:"items"`
}

type ConfigMap struct {
	ApiVersion string            `json:"apiVersion"`
	Data       map[string]string `json:"data"`
	Kind       string            `json:"kind"`
	Metadata   Metadata          `json:"metadata"`
}

type SecretList struct {
	Items []Secret `json:"items"`
}

type Secret struct {
	ApiVersion string            `json:"apiVersion"`
	Data       map[string]string `json:"data"`
	Kind       string            `json:"kind"`
	Metadata   Metadata          `json:"metadata"`
	Type       string            `json:"type"`
}

type Metadata struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

func getConfigMap(namespace, name string) (*ConfigMap, error) {
	u := fmt.Sprintf("http://127.0.0.1:8001/api/v1/namespaces/%s/configmaps/%s", namespace, name)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("non 200 response code")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	var cm ConfigMap
	if err := json.Unmarshal(data, &cm); err != nil {
		return nil, err
	}

	return &cm, nil
}

func getConfigMaps(namespace string) (*ConfigMapList, error) {
	u := fmt.Sprintf("http://127.0.0.1:8001/api/v1/namespaces/%s/configmaps?labelSelector=konfd.io/template=true",
		namespace)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("non 200 response code")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	var cl ConfigMapList
	err = json.Unmarshal(data, &cl)
	if err != nil {
		return nil, err
	}
	return &cl, nil
}

func createConfigMap(namespace, name, key, value string) error {
	c := ConfigMap{
		ApiVersion: "v1",
		Data:       make(map[string]string),
		Kind:       "ConfigMap",
		Metadata: Metadata{
			Name:      name,
			Namespace: namespace,
		},
	}
	c.Data[key] = value

	body, err := json.MarshalIndent(&c, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding configmap %s: %v", name, err)
	}

	u := fmt.Sprintf("http://127.0.0.1:8001/api/v1/namespaces/%s/configmaps", namespace)
	resp, err := http.Post(u, "", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("error creating configmap %s: %v", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("error creating configmap %s; got HTTP %v status code", name, resp.StatusCode)
	}

	return nil
}

func getNamespaces() (*NamespaceList, error) {
	resp, err := http.Get("http://127.0.0.1:8001/api/v1/namespaces")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("non 200 response code")
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	var nl NamespaceList
	err = json.Unmarshal(data, &nl)
	if err != nil {
		return nil, err
	}
	return &nl, nil
}
