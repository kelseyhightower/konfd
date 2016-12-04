package main

import (
	"bytes"
	"html/template"
	"log"
)

func main() {
	namespaces, err := getNamespaces()
	if err != nil {
		log.Fatal(err)
	}

	for _, namespace := range namespaces.Items {
		namespaceName := namespace.Metadata.Name

		configmaps, err := getConfigMaps(namespaceName)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, configmap := range configmaps.Items {
			ts, ok := configmap.Data["template"]
			if !ok {
				log.Println("missing template key")
				continue
			}
			t := template.New(configmap.Metadata.Name)
			tp := &TemplateProcessor{
				namespace:  namespaceName,
				configMaps: make(map[string]*ConfigMap),
				secrets:    make(map[string]*Secret),
				templates:  make(map[string]*ConfigMap),
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

			var value bytes.Buffer
			err = t.Execute(&value, nil)
			if err != nil {
				log.Println("error executing template: %v", err)
				continue
			}

			annotations := configmap.Metadata.Annotations
			name := annotations["konfd.io/name"]
			key := annotations["konfd.io/key"]

			err = createConfigMap(namespaceName, name, key, value.String())
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}
