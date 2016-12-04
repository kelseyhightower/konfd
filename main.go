package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	resp, err := http.Get("http://127.0.0.1:8001/api/v1/namespaces")
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatal("non 200 response code")
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	var ns NamespaceList
	err = json.Unmarshal(data, &ns)
	if err != nil {
		log.Fatal(err)
	}
	for _, n := range ns.Items {
		u := fmt.Sprintf("http://127.0.0.1:8001/api/v1/namespaces/%s/configmaps?labelSelector=konfd.io/template=true",
			n.Metadata.Name)
		resp, err := http.Get(u)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode != 200 {
			log.Fatal("non 200 response code")
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()

		var cs ConfigMapList
		err = json.Unmarshal(data, &cs)
		if err != nil {
			log.Fatal(err)
		}

		for _, cm := range cs.Items {
			ts, ok := cm.Data["template"]
			if !ok {
				log.Println("missing template key")
				continue
			}
			t := template.New(cm.Metadata.Name)
			tp := &TemplateProcessor{
				namespace:  n.Metadata.Name,
				configMaps: make(map[string]ConfigMap),
				secrets:    make(map[string]Secret),
				templates:  make(map[string]ConfigMap),
			}

			t.Funcs(template.FuncMap{
				"configmap": tp.configmap,
				"secret":    tp.secret,
			})

			t, err := t.Parse(ts)
			if err != nil {
				log.Println("error parsing template: %v", err)
				continue
			}

			var buf bytes.Buffer
			err = t.Execute(&buf, nil)
			if err != nil {
				log.Println("error executing template: %v", err)
				continue
			}

			annotations := cm.Metadata.Annotations
			cm2 := ConfigMap{
				ApiVersion: "v1",
				Data:       make(map[string]string),
				Kind:       "ConfigMap",
				Metadata: Metadata{
					Name:      annotations["konfd.io/name"],
					Namespace: n.Metadata.Name,
				},
			}

			cm2.Data[annotations["konfd.io/key"]] = buf.String()
			jsonEncoded, err := json.MarshalIndent(&cm2, "", "  ")
			if err != nil {
				log.Printf("error encoding object: %v", err)
				continue
			}

			configmapURL := fmt.Sprintf("http://127.0.0.1:8001/api/v1/namespaces/%s/configmaps", n.Metadata.Name)

			log.Println(string(jsonEncoded))
			resp, err := http.Post(configmapURL, "", bytes.NewReader(jsonEncoded))
			if err != nil {
				log.Printf("error posting object: %v", err)
				continue
			}
			if resp.StatusCode != 201 {
				log.Printf("error posting object: non 200 error: %v", resp.StatusCode)
				continue
			}
		}
	}
}
