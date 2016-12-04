package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type TemplateProcessor struct {
	configMaps map[string]ConfigMap
	namespace  string
	secrets    map[string]Secret
	templates  map[string]ConfigMap
}

func (t *TemplateProcessor) configmap(name, key string) (string, error) {
	_, ok := t.configMaps[name]
	if !ok {
		u := fmt.Sprintf("http://127.0.0.1:8001/api/v1/namespaces/%s/configmaps/%s", t.namespace, name)
		resp, err := http.Get(u)
		if err != nil {
			return "", err
		}

		if resp.StatusCode != 200 {
			return "", errors.New("non 200 response code")
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		resp.Body.Close()

		var cm ConfigMap
		err = json.Unmarshal(data, &cm)
		if err != nil {
			return "", err
		}
		t.configMaps[name] = cm
	}
	value, ok := t.configMaps[name].Data[key]
	if !ok {
		return "", errors.New("missing key " + key)
	}

	return value, nil
}

func (t *TemplateProcessor) secret(name, key string) (string, error) {
	_, ok := t.secrets[name]
	if !ok {
		u := fmt.Sprintf("http://127.0.0.1:8001/api/v1/namespaces/%s/secrets/%s", t.namespace, name)
		resp, err := http.Get(u)
		if err != nil {
			return "", err
		}

		if resp.StatusCode != 200 {
			return "", errors.New("non 200 response code")
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		resp.Body.Close()

		var s Secret
		err = json.Unmarshal(data, &s)
		if err != nil {
			return "", err
		}
		t.secrets[name] = s
	}
	value, ok := t.secrets[name].Data[key]
	if !ok {
		return "", errors.New("missing key " + key)
	}
	d, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}

	return string(d), nil
}

func (t *TemplateProcessor) sync() error {
	return nil
}
