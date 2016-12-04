package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"html/template"
	"log"
)

type TemplateProcessor struct {
	configMaps map[string]*ConfigMap
	namespace  string
	secrets    map[string]*Secret
	templates  map[string]*ConfigMap
}

func NewTemplateProcessor(namespace string) *TemplateProcessor {
	return &TemplateProcessor{
		namespace:  namespace,
		configMaps: make(map[string]*ConfigMap),
		secrets:    make(map[string]*Secret),
		templates:  make(map[string]*ConfigMap),
	}
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

func (tp *TemplateProcessor) sync() {
	configmaps, err := getConfigMaps(tp.namespace)
	if err != nil {
		log.Println(err)
	}

	for _, configmap := range configmaps.Items {
		ts, ok := configmap.Data["template"]
		if !ok {
			log.Println("missing template key")
			continue
		}

		t := template.New(configmap.Metadata.Name)
		t.Funcs(template.FuncMap{
			"configmap": tp.configmap,
			"secret":    tp.secret,
		})

		t, err := t.Parse(ts)
		if err != nil {
			log.Println("error parsing template: %v", err)
			continue
		}

		var value bytes.Buffer
		err = t.Execute(&value, nil)
		if err != nil {
			log.Println("error executing template: %v", err)
			continue
		}

		annotations := configmap.Metadata.Annotations
		name := annotations["konfd.io/name"]
		key := annotations["konfd.io/key"]

		err = createConfigMap(tp.namespace, name, key, value.String())
		if err != nil {
			log.Println(err)
			continue
		}
	}
}
