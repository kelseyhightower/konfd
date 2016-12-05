// Copyright 2016 Google Inc. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
//
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var ErrNotExist = errors.New("object does not exist")

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
	StringData map[string]string `json:"stringData,omitempty"`
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

	if resp.StatusCode == 404 {
		return nil, ErrNotExist
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

func getSecret(namespace, name string) (*Secret, error) {
	u := fmt.Sprintf("http://127.0.0.1:8001/api/v1/namespaces/%s/secrets/%s", namespace, name)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, ErrNotExist
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("non 200 response code")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	var s Secret
	err = json.Unmarshal(data, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
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

func newConfigMap(namespace, name, key, value string) *ConfigMap {
	c := &ConfigMap{
		ApiVersion: "v1",
		Data:       make(map[string]string),
		Kind:       "ConfigMap",
		Metadata: Metadata{
			Name:      name,
			Namespace: namespace,
		},
	}
	c.Data[key] = value
	return c
}

func newSecret(namespace, name, key, value string) *Secret {
	s := &Secret{
		ApiVersion: "v1",
		Data:       make(map[string]string),
		Kind:       "Secret",
		Metadata: Metadata{
			Name:      name,
			Namespace: namespace,
		},
		Type: "Opaque",
	}
	s.Data[key] = base64.StdEncoding.EncodeToString([]byte(value))
	return s
}

func createConfigMap(c *ConfigMap) error {
	body, err := json.MarshalIndent(&c, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding configmap %s: %v", c.Metadata.Name, err)
	}

	u := fmt.Sprintf("http://127.0.0.1:8001/api/v1/namespaces/%s/configmaps", c.Metadata.Namespace)
	resp, err := http.Post(u, "", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("error creating configmap %s: %v", c.Metadata.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("error creating configmap %s; got HTTP %v status code", c.Metadata.Name, resp.StatusCode)
	}

	return nil
}

func createSecret(s *Secret) error {
	body, err := json.MarshalIndent(&s, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding secret %s: %v", s.Metadata.Name, err)
	}

	u := fmt.Sprintf("http://127.0.0.1:8001/api/v1/namespaces/%s/secrets", s.Metadata.Namespace)
	resp, err := http.Post(u, "", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("error creating secrets %s: %v", s.Metadata.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("error creating secrets %s; got HTTP %v status code", s.Metadata.Name, resp.StatusCode)
	}

	return nil
}

func updateConfigMap(c *ConfigMap) error {
	body, err := json.MarshalIndent(&c, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding configmap %s: %v", c.Metadata.Name, err)
	}

	u := fmt.Sprintf("http://127.0.0.1:8001/api/v1/namespaces/%s/configmaps/%s", c.Metadata.Namespace, c.Metadata.Name)
	request, err := http.NewRequest(http.MethodPut, u, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("error updating configmap %s: %v", c.Metadata.Name, err)
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("error updating configmap %s: %v", c.Metadata.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("error updating configmap %s; got HTTP %v status code", c.Metadata.Name, resp.StatusCode)
	}

	return nil
}

func updateSecret(s *Secret) error {
	body, err := json.MarshalIndent(&s, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding secret %s: %v", s.Metadata.Name, err)
	}

	u := fmt.Sprintf("http://127.0.0.1:8001/api/v1/namespaces/%s/secrets/%s", s.Metadata.Namespace, s.Metadata.Name)
	request, err := http.NewRequest(http.MethodPut, u, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("error updating secret %s: %v", s.Metadata.Name, err)
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("error updating secret %s: %v", s.Metadata.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("error updating secret %s; got HTTP %v status code", s.Metadata.Name, resp.StatusCode)
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

func waitForKubernetesProxy() {
	for {
		resp, err := http.Get("http://127.0.0.1:8001/api")
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second)
			continue
		}
		resp.Body.Close()
		return
	}
}
