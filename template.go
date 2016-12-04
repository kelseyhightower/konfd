package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
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

func (tp *TemplateProcessor) sync(name string) error {
	cm, err := getConfigMap(tp.namespace, name)
	if err != nil {
		return err
	}

	return tp.processConfigMapTemplate(cm)
}

func (tp *TemplateProcessor) syncAll() {
	configmaps, err := getConfigMaps(tp.namespace)
	if err != nil {
		log.Println(err)
		return
	}

	for _, configmap := range configmaps.Items {
		if err := tp.processConfigMapTemplate(&configmap); err != nil {
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

	c := newConfigMap(tp.namespace, name, key, value.String())

	if tp.noop {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		err := encoder.Encode(&c)
		if err != nil {
			return fmt.Errorf("error encoding configmap %s: %v", name, err)
		}
		return nil
	}

	return createConfigMap(c)
}
