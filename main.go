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
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	configmaps   stringSlice
	namespaces   stringSlice
	noop         bool
	onetime      bool
	syncInterval time.Duration
)

func main() {
	flag.Var(&namespaces, "namespace", "the namespace to process.")
	flag.Var(&configmaps, "configmap", "the configmap to process.")
	flag.BoolVar(&noop, "noop", false, "print processed configmaps and secrets and do not submit them to the cluster.")
	flag.BoolVar(&onetime, "onetime", false, "run one time and exit.")
	flag.DurationVar(&syncInterval, "sync-interval", 60, "the number of seconds between template processing.")
	flag.Parse()

	if onetime {
		process(namespaces, configmaps, noop)
		os.Exit(0)
	}

	log.Println("Starting konfd...")
	var wg sync.WaitGroup
	done := make(chan struct{})

	go func() {
		wg.Add(1)
		for {
			process(namespaces, configmaps, noop)
			log.Printf("Syncing templates complete. Next sync in %d seconds.", syncInterval)
			select {
			case <-time.After(syncInterval * time.Second):
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
