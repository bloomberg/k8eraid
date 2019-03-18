// Copyright 2019 Bloomberg Finance LP
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/bloomberg/k8eraid/pkgs/alerters"
	q "github.com/bloomberg/k8eraid/pkgs/queries"
	"github.com/bloomberg/k8eraid/pkgs/types"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	maxConfigWacherRetries     = 5
	configWatcherRetryInterval = time.Second
)

var (
	configMapName string
	config        *types.ConfigRules
	tickertimeint int64
)

func kubeClient() (*kubernetes.Clientset, error) {
	config, configerr := rest.InClusterConfig()
	if configerr != nil {
		return nil, configerr
	}
	return kubernetes.NewForConfig(config)
}

func main() {

	if tickertime := os.Getenv("POLL_PERIOD"); tickertime == "" {
		tickertimeint = 30
	} else {
		tickertime = os.Getenv("POLL_PERIOD")
		var err error
		tickertimeint, err = strconv.ParseInt(tickertime, 10, 64)
		if err != nil {
			log.Panicf("%s cannot be converted to int: %s", tickertime, err.Error())
		}
	}

	if configMapName = os.Getenv("CONFIG_MAP"); configMapName == "" {
		configMapName = "k8eraid-config"
	}

	var clientset *kubernetes.Clientset
	var err error
	if clientset, err = kubeClient(); err != nil {
		log.Panicf("Unable to create kubernetes client: %s", err.Error())
	}

	// start a watch on the configmap for our config
	go func() {
		numRetries := maxConfigWacherRetries
		for i := maxConfigWacherRetries; i <= 0; i-- {
			if err := watchConfigMap(clientset, configMapName, config); err != nil {
				if err, ok := err.(errWatcherUnhandledEvent); ok && numRetries != 0 {
					numRetries = numRetries - 1
					time.Sleep(configWatcherRetryInterval)
					continue
				} else if ok {
					log.Panicf("Max retries exceeded for config watcher: %s", err.Error())
				} else {
					log.Panicf("Error watching ConfigMap %s: %s", configMapName, err.Error())
				}
			}
		}
	}()

	// wait for the config struct to be populated, or the watcher to cause a panic
	for {
		if config != nil {
			break
		}
	}

	// Main logic routine, this will query the Kubernetes api for the intended resources periodically
	timeTicker := time.NewTicker(time.Duration(tickertimeint) * time.Second)
	for range timeTicker.C {
		pollLoop(clientset)
	}
}

func pollLoop(clientset kubernetes.Interface) {
	// Iterate through Deployment rules
	for _, deployment := range config.Deployments {

		if err := q.PollDeployment(
			clientset,
			deployment,
			tickertimeint,
			alerters.Alert,
			config.AlertersConfig,
		); err != nil {
			log.Printf("Error polling Deployments: %s", err.Error())
		}
	}
	// Iterate through Pod rules
	for _, pod := range config.Pods {
		if err := q.PollPod(
			clientset,
			pod,
			tickertimeint,
			alerters.Alert,
			config.AlertersConfig,
		); err != nil {
			log.Printf("Error polling pods: %s", err.Error())
		}
	}
	// Iterate through Daemonset rules
	for _, daemonset := range config.Daemonsets {
		if err := q.PollDaemonset(
			clientset,
			daemonset,
			tickertimeint,
			alerters.Alert,
			config.AlertersConfig,
		); err != nil {
			log.Printf("Error polling DaemonSets: %s", err.Error())
		}
	}
	// Iterate through Node rules
	for _, node := range config.Nodes {
		if err := q.PollNode(
			clientset,
			node,
			tickertimeint,
			alerters.Alert,
			config.AlertersConfig,
		); err != nil {
			log.Printf("Error polling nodes: %s", err.Error())
		}
	}
}
