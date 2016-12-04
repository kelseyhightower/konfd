package main

import (
	"flag"
	"log"
	"os"
	"time"
)

func main() {
	namespace := flag.String("namespace", "default", "The Kubernetes namespace")
	name := flag.String("configmap-name", "", "The configmap name")
	noop := flag.Bool("noop", false, "Process template configmap and print to standard out")
	syncInterval := flag.Int64("sync-interval", 60, "Sync interval in seconds")
	flag.Parse()

	if *name != "" {
		tp := NewTemplateProcessor(*namespace)
		tp.setNoop(*noop)
		if err := tp.sync(*name); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	log.Println("Starting konfd ...")
	for {
		log.Println("Syncing templates...")
		namespaces, err := getNamespaces()
		if err != nil {
			log.Println(err)
			continue
		}

		for _, namespace := range namespaces.Items {
			namespaceName := namespace.Metadata.Name
			tp := NewTemplateProcessor(namespaceName)
			tp.syncAll()
		}

		log.Printf("Syncing templates complete. Next sync in %d seconds", *syncInterval)
		<-time.After(time.Duration(*syncInterval) * time.Second)
	}
}
