package main

import "log"

func main() {
	namespaces, err := getNamespaces()
	if err != nil {
		log.Fatal(err)
	}

	for _, namespace := range namespaces.Items {
		namespaceName := namespace.Metadata.Name
		tp := NewTemplateProcessor(namespaceName)
		tp.sync()
	}
}
