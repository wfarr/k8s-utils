package main

import (
  "encoding/json"
  "flag"
  "time"

  "k8s.io/kubernetes/pkg/api"
  client "k8s.io/kubernetes/pkg/client/unversioned"
  "k8s.io/kubernetes/pkg/fields"
  "k8s.io/kubernetes/pkg/labels"
  log "github.com/Sirupsen/logrus"
)

func main() {
  interval := flag.Duration("interval", 30 * time.Second, "time in seconds to sleep between GCs")
  flag.Parse()

  for {
    gc()
    log.Info("sleeping for ", *interval, "s")
    time.Sleep(*interval)
  }
}

func gc() {
  log.Info("Starting GC run...")

  kubeClient, err := client.NewInCluster()
  if err != nil {
    log.Fatal(err)
  }

  listAll := api.ListOptions{
    LabelSelector: labels.Everything(),
    FieldSelector: fields.Everything(),
  }

  deleteOpts := &api.DeleteOptions{}

  namespaces, err := kubeClient.Namespaces().List(listAll)
  if err != nil {
    log.Error(err)
  }

  log.Debug("iterating over ", len(namespaces.Items), " namespaces")
  for _, namespace := range namespaces.Items {
    log.Debug("in namespace:", namespace.Name)

    log.Debug("getting pods")
    pods, err := kubeClient.Pods(namespace.Name).List(listAll)
    if err != nil {
      log.Error(err)
    }
    log.Debug("got pods")

    log.Debug("iterating over ", len(pods.Items), " pods")
    for _, pod := range pods.Items {
      creatorRefJson, found := pod.ObjectMeta.Annotations["kubernetes.io/created-by"]
      if !found {
        log.Debug("no created-by ref, skipping pod:", pod.Name)
        continue
      }

      log.Debug("have created-by ref for pod:", pod.Name)
      log.Debug("created-by:", creatorRefJson)

      log.Debug("parsing json")
      var sr api.SerializedReference
	    err := json.Unmarshal([]byte(creatorRefJson), &sr)
      if err != nil {
        log.Error(err)
      }
      log.Debug("parsed json")

      log.Debug("this pod has the following reference:", sr.Reference)
      if sr.Reference.Kind == "Job" {
        log.Debug("found job pod:", pod)

        switch(pod.Status.Phase) {
        case api.PodSucceeded:
          log.Info("cleaning up pod that exited successfully:", pod.Name)
          err := kubeClient.Pods(namespace.Name).Delete(pod.Name, deleteOpts)
          if err != nil {
            log.Error(err)
          }
        case api.PodFailed:
          log.Info("cleaning up pod that exited with failures:", pod.Name)
          err := kubeClient.Pods(namespace.Name).Delete(pod.Name, deleteOpts)
          if err != nil {
            log.Error(err)
          }
        default:
          log.Info("skipping pod:", pod.Name, "state:", pod.Status.Phase)
        }
      }
    }
  }
}
