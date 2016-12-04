// Copyright 2016 Google Inc. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	log.Println("Starting konfd...")
	var wg sync.WaitGroup
	done := make(chan struct{})

	go func() {
		wg.Add(1)
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

			log.Printf("Syncing templates complete. Next sync in %d seconds.", *syncInterval)
			select {
			case <-time.After(time.Duration(*syncInterval) * time.Second):
			case <-done:
				wg.Done()
				return
			}
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan
	log.Printf("Shutdown signal received, exiting...")
	close(done)
	wg.Wait()

	os.Exit(0)
}
