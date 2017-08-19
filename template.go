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
	"text/template"
	"log"
	"os"
)

type TemplateProcessor struct {
	configMaps map[string]*ConfigMap
	namespace  string
	secrets    map[string]*Secret
	templates  map[string]*ConfigMap
	noop       bool
}

func NewTemplateProcessor(namespace string) *TemplateProcessor {
	return &TemplateProcessor{
		namespace:  namespace,
		configMaps: make(map[string]*ConfigMap),
		secrets:    make(map[string]*Secret),
		templates:  make(map[string]*ConfigMap),
	}
}

func (tp *TemplateProcessor) setNoop(noop bool) {
	tp.noop = noop
}

func (tp *TemplateProcessor) configmap(name, key string) (string, error) {
	// Check if the config map has already been fetched for this
	// namespace. If not, retrieve the config map and cache it for
	// future use.
	_, ok := tp.configMaps[name]
	if !ok {
		cm, err := getConfigMap(tp.namespace, name)
		if err != nil {
			return "", err
		}
		tp.configMaps[name] = cm
	}

	v, ok := tp.configMaps[name].Data[key]
	if !ok {
		return "", errors.New("missing key " + key)
	}

	return v, nil
}

func (tp *TemplateProcessor) secret(name, key string) (string, error) {
	_, ok := tp.secrets[name]
	if !ok {
		s, err := getSecret(tp.namespace, name)
		if err != nil {
			return "", err
		}
		tp.secrets[name] = s
	}

	v, ok := tp.secrets[name].Data[key]
	if !ok {
		return "", errors.New("missing key " + key)
	}

	d, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return "", err
	}

	return string(d), nil
}

func process(namespaces, configmaps []string, noop bool) {
	if len(namespaces) == 0 {
		ns, err := getNamespaces()
		if err != nil {
			log.Println(err)
			return
		}
		for _, n := range ns.Items {
			namespaces = append(namespaces, n.Metadata.Name)
		}
	}

	for _, n := range namespaces {
		tp := NewTemplateProcessor(n)
		tp.setNoop(noop)
		tp.sync(configmaps)
	}
}

func (tp *TemplateProcessor) sync(configmaps []string) {
	var cms []*ConfigMap

	if len(configmaps) == 0 {
		cmList, err := getConfigMaps(tp.namespace)
		if err != nil {
			log.Println(err)
			return
		}
		for i := range cmList.Items {
			cms = append(cms, &cmList.Items[i])
		}
	}

	for _, c := range configmaps {
		cm, err := getConfigMap(tp.namespace, c)
		if err != nil {
			log.Println(err)
			continue
		}
		cms = append(cms, cm)
	}

	for _, c := range cms {
		if err := tp.processConfigMapTemplate(c); err != nil {
			log.Println(err)
			continue
		}
	}
}

func (tp *TemplateProcessor) processConfigMapTemplate(configmap *ConfigMap) error {
	ts, ok := configmap.Data["template"]
	if !ok {
		return errors.New("missing template key")
	}

	t := template.New(configmap.Metadata.Name)
	t.Funcs(template.FuncMap{
		"configmap": tp.configmap,
		"secret":    tp.secret,
	})

	t, err := t.Parse(ts)
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	var value bytes.Buffer
	err = t.Execute(&value, nil)
	if err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	annotations := configmap.Metadata.Annotations
	name := annotations["konfd.io/name"]
	key := annotations["konfd.io/key"]
	kind := annotations["konfd.io/kind"]

	switch kind {
	case "configmap":
		return tp.processConfigMap(tp.namespace, name, key, value.String())
	case "secret":
		return tp.processSecret(tp.namespace, name, key, value.String())
	}
	return nil
}

func (tp *TemplateProcessor) processConfigMap(namespace, name, key, value string) error {
	cm := newConfigMap(namespace, name, key, value)

	ccm, err := getConfigMap(namespace, name)
	if err == ErrNotExist {
		if tp.noop {
			return printObject(cm)
		}
		return createConfigMap(cm)
	}

	if err != nil {
		return err
	}

	if ccm.Data[key] != cm.Data[key] {
		log.Printf("%s configmap out of sync; syncing...", name)
		ccm.Data[key] = cm.Data[key]
	}

	if tp.noop {
		return printObject(ccm)
	}
	return updateConfigMap(ccm)
}

func (tp *TemplateProcessor) processSecret(namespace, name, key, value string) error {
	s := newSecret(namespace, name, key, value)

	currentSecret, err := getSecret(namespace, name)
	if err == ErrNotExist {
		if tp.noop {
			return printObject(s)
		}
		return createSecret(s)
	}

	if err != nil {
		return err
	}

	if currentSecret.Data[key] != s.Data[key] {
		log.Printf("%s secret out of sync; syncing...", name)
		currentSecret.Data[key] = s.Data[key]
	}

	if tp.noop {
		return printObject(currentSecret)
	}
	return updateSecret(currentSecret)
}

func printObject(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(&v)
	if err != nil {
		return fmt.Errorf("error encoding object: %v", err)
	}
	return nil
}
