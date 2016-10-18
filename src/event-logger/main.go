package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/Sirupsen/logrus"
)

func main() {
	log.Info("starting event logger...")

	apiServer := flag.String("api-server", "http://127.0.0.1:8001", "kubernetes api server")
	apiVersion := flag.String("api-version", "v1", "the kubernetes API version (should not have to alter thus except for custom resources)")
	inCluster := flag.Bool("in-cluster", false, "invoked from inside a k8s cluster")
	resources := flag.String("resources", "pods", "the kind of resources to log events for")
	flag.Parse()

	apiHost = *apiServer
	apiPrefix = *apiVersion
	usePodServiceAccount = *inCluster
	resourceKind = *resources

	done := make(chan struct{})
	var wg sync.WaitGroup

	log.Info("watching events for " + resourceKind)
	processResourceEvents(done, &wg)
	wg.Add(1)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-signals:
			log.Info("received signal, shutting down event logger...")
			close(done)
			wg.Wait()
			os.Exit(0)
		}
	}
}

func processResourceEvents(done chan struct{}, wg *sync.WaitGroup) {
	events, errs := watchResourceEvents()
	go func() {
		for {
			select {
			case event := <-events:
				err := processResourceEvent(event)
				if err != nil {
					log.Error(err)
				}
			case err := <-errs:
				log.Error(err)
			case <-done:
				wg.Done()
				log.Info("stopping event processor")
				return
			}
		}
	}()
}

func processResourceEvent(event ResourceEvent) error {
	log.Info("received event: ", fmt.Sprintf("%#v", event))
	return nil
}
