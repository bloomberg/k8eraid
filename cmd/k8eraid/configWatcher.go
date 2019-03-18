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
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/bloomberg/k8eraid/pkgs/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type errWatcherUnhandledEvent struct {
	Type watch.EventType
}

func (e errWatcherUnhandledEvent) Error() string {
	return fmt.Sprintf("ConfigMap watcher got event of type %s, cannot continue", e.Type)
}

func watchConfigMap(client kubernetes.Interface, configMapName string, config *types.ConfigRules) error {
	opts := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", configMapName),
		Watch:         true,
	}
	if watcher, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Watch(opts); err == nil {
		for e := range watcher.ResultChan() {
			if err := eventReceived(e, config); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("unable to watch ConfigMap: %s", err.Error())
	}
	return errors.New("ConfigMap watcher ended")
}

func eventReceived(e watch.Event, config *types.ConfigRules) error {
	if e.Type == watch.Added || e.Type == watch.Modified {
		log.Printf("ConfigMap %s changed, updating config", configMapName)
		if configMap, ok := e.Object.(*corev1.ConfigMap); ok {
			if configJSON, ok := configMap.Data["config.json"]; ok {
				if err := json.Unmarshal([]byte(configJSON), config); err != nil {
					return fmt.Errorf("unable to parse new config from %s: %s", configMapName, err.Error())
				}
				for _, pod := range config.Pods {
					log.Println("Pod rule found for: ", pod.Name)
				}

				for _, daemonSet := range config.Daemonsets {
					log.Println("Daemonset rule found for: ", daemonSet.Name)
				}

				for _, node := range config.Nodes {
					log.Println("Node rule found for: ", node.Name)
				}
			} else {
				return fmt.Errorf("ConfigMap %s missing config.json key", configMapName)
			}
		} else {
			return fmt.Errorf("unable to coerce event object of kind %s to ConfigMap", e.Object.GetObjectKind())
		}
	} else {
		return &errWatcherUnhandledEvent{Type: e.Type}
	}
	return nil
}
