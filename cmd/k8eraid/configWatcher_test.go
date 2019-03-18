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
	"io/ioutil"
	"strings"
	"testing"

	"github.com/bloomberg/k8eraid/pkgs/types"

	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

func Test_EventReceived_ok(t *testing.T) {
	config := types.ConfigRules{}
	var configMap corev1.ConfigMap
	if data, readerr := ioutil.ReadFile("../../examples/k8eraid-configmap.yml"); readerr == nil {
		if yamlerr := yaml.Unmarshal(data, &configMap); yamlerr != nil {
			panic(yamlerr)
		}
	} else {
		panic(readerr)
	}
	event := watch.Event{
		Type:   watch.Added,
		Object: &configMap,
	}
	if err := eventReceived(event, &config); err != nil {
		t.Errorf("eventReceived returned an unexpected error: %s", err.Error())
	}
	// Test deployment config loading
	if config.Deployments[0].Name != "heapster" {
		t.Errorf(
			"Config had unexpected result for deployment name, got: %s, expected: %s",
			config.Deployments[0].Name,
			"kube-dns",
		)
	}
	if config.Deployments[0].DepFilter != "kube-system" {
		t.Errorf(
			"Config had unexpected result for deployment filter, got: %s, expected: %s",
			config.Deployments[0].DepFilter,
			"kube-system",
		)
	}
	if config.Deployments[0].ReportStatus.MinReplicas != 1 {
		t.Errorf(
			"Config had unexpected result for deployment minimum replica count, got %d, expected: %d",
			config.Deployments[0].ReportStatus.MinReplicas,
			1,
		)
	}
	if config.Deployments[0].ReportStatus.PendingThreshold != 1 {
		t.Errorf(
			"Config had unexpected result for deployment pending threshold, got %d, expected: %d",
			config.Deployments[0].ReportStatus.PendingThreshold,
			1,
		)
	}
	if config.Deployments[0].AlerterType != "smtp" {
		t.Errorf(
			"Config had unexpected result for deployment alert type, got: %s, expected: %s",
			config.Deployments[0].AlerterType,
			"stderr",
		)
	}

	// Test pod config loading
	if config.Pods[0].Name != "*" {
		t.Errorf("Config had unexpected result for pod name, got: %s, expected: %s", config.Pods[0].Name, "*")
	}
	if config.Pods[0].PodFilterNamespace != "kube-system" {
		t.Errorf(
			"Config had unexpected result for pod namespace filter, got: %s, expected: %s",
			config.Pods[0].PodFilterNamespace,
			"kube-system",
		)
	}
	if config.Pods[0].PodFilterLabel != "" {
		t.Errorf(
			"Config had unexpected result for pod label filter, got: %s, expected: %s",
			config.Pods[0].PodFilterLabel,
			"",
		)
	}
	if config.Pods[0].ReportStatus.MinPods != 1 {
		t.Errorf(
			"Config had unexpected result for pod minimum count got: %d, expected %d",
			config.Pods[0].ReportStatus.MinPods,
			1,
		)
	}
	if config.Pods[0].ReportStatus.PodRestarts != true {
		t.Errorf("Config had unexpected result for pod restart check, got: %t, expected %t",
			config.Pods[0].ReportStatus.PodRestarts,
			true,
		)
	}
	if config.Pods[0].ReportStatus.FailedScheduling != true {
		t.Errorf("Config had unexpected result for pod scedulting failed check, got: %t, expected %t",
			config.Pods[0].ReportStatus.FailedScheduling,
			true,
		)
	}
	if config.Pods[0].ReportStatus.StuckTerminating != true {
		t.Errorf("Config had unexpected result for pod stuck terminating check, got: %t, expected %t",
			config.Pods[0].ReportStatus.StuckTerminating,
			true,
		)
	}
	if config.Pods[0].ReportStatus.PendingThreshold != 1 {
		t.Errorf("Config had unexpected result for pod pending threshold, got: %d, expected %d",
			config.Pods[0].ReportStatus.PendingThreshold,
			1,
		)
	}
	if config.Pods[0].AlerterType != "smtp" {
		t.Errorf("Config had unexpected result for pod alerter type, got: %s, expected: %s",
			config.Pods[0].AlerterType,
			"smtp",
		)
	}
	if config.Pods[0].AlerterName != "example-email" {
		t.Errorf("Config had unexpected result for pod alerter name, got: %s, expected: %s",
			config.Pods[0].AlerterName,
			"example-email",
		)
	}

	// Test daemonset config loading
	if config.Daemonsets[0].Name != "nginx-ingress-controller" {
		t.Errorf("Config had unexpected result for daemonset name, got: %s, expected: %s",
			config.Daemonsets[0].Name,
			"nginx-ingress-controller",
		)
	}
	if config.Daemonsets[0].DaemonFilter != "kube-system" {
		t.Errorf("Config had unexpected result for daemonset filter, got: %s, expected: %s",
			config.Daemonsets[0].DaemonFilter,
			"kube-system",
		)
	}
	if config.Daemonsets[0].ReportStatus.CheckReplicas != true {
		t.Errorf("Config had unexpected result for daemonset check replica reporting, got: %t, expected %t",
			config.Daemonsets[0].ReportStatus.CheckReplicas,
			true,
		)
	}
	if config.Daemonsets[0].ReportStatus.FailedScheduling != true {
		t.Errorf("Config had unexpected result for daemonset failed scheduling reporting, got: %t, expected %t",
			config.Daemonsets[0].ReportStatus.FailedScheduling,
			true,
		)
	}
	if config.Daemonsets[0].ReportStatus.PendingThreshold != 1 {
		t.Errorf("Config had unexpected result for daemonset pending threshold, got: %d, expected: %d",
			config.Daemonsets[0].ReportStatus.PendingThreshold,
			1,
		)
	}
	if config.Daemonsets[0].AlerterType != "smtp" {
		t.Errorf("Config had unexpected result for daemonset alert type, got: %s, expected: %s",
			config.Daemonsets[0].AlerterType,
			"smtp",
		)
	}
	if config.Daemonsets[0].AlerterName != "example-email" {
		t.Errorf("Config had unexpected result for daemonset alert name, got: %s, expected: %s",
			config.Daemonsets[0].AlerterName,
			"example-email",
		)
	}

	// Test node config loading
	if config.Nodes[0].Name != "*" {
		t.Errorf("Config had unexpected result for node name, got: %s, expected: %s",
			config.Nodes[0].Name,
			"*",
		)
	}
	if config.Nodes[0].NodeFilter != "" {
		t.Errorf(
			"Config had unexpected result for node filter, got: %s, expected: %s",
			config.Nodes[0].NodeFilter,
			"",
		)
	}
	if config.Nodes[0].ReportStatus.NodeReady != true {
		t.Errorf("Config had unexpected result for node readiness check, got: %t, expected %t",
			config.Nodes[0].ReportStatus.NodeReady,
			true,
		)
	}
	if config.Nodes[0].ReportStatus.NodeOutOfDisk != true {
		t.Errorf("Config had unexpected result for node out of disk check, got: %t, expected %t",
			config.Nodes[0].ReportStatus.NodeOutOfDisk,
			true,
		)
	}
	if config.Nodes[0].ReportStatus.NodeMemoryPressure != true {
		t.Errorf("Config had unexpected result for node memory pressure check, got: %t, expected %t",
			config.Nodes[0].ReportStatus.NodeMemoryPressure,
			true,
		)
	}
	if config.Nodes[0].ReportStatus.NodeDiskPressure != true {
		t.Errorf("Config had unexpected result for node disk pressure check, got: %t, expected %t",
			config.Nodes[0].ReportStatus.NodeDiskPressure,
			true,
		)
	}
	if config.Nodes[0].ReportStatus.MinNodes != 3 {
		t.Errorf("Config had unexpected result for node count minimum, got: %d, expected: %d",
			config.Nodes[0].ReportStatus.MinNodes,
			3,
		)
	}
	if config.Nodes[0].ReportStatus.PendingThreshold != 60 {
		t.Errorf("Config had unexpected result for node pending threshold, got: %d, expected: %d",
			config.Nodes[0].ReportStatus.PendingThreshold,
			60,
		)
	}
	if config.Nodes[0].AlerterType != "smtp" {
		t.Errorf("Config had unexpected result for node alerter type, got: %s, expected: %s",
			config.Nodes[0].AlerterType,
			"smtp",
		)
	}
	if config.Nodes[0].AlerterName != "example-email" {
		t.Errorf("Config had unexpected result for node alerter name, got: %s, expected: %s",
			config.Nodes[0].AlerterName,
			"example-email",
		)
	}
}

// These tests are super naive and won't fail if no error is raised.
func Test_EventReceived_err(t *testing.T) {
	config := types.ConfigRules{}
	tests := []struct {
		name      string
		configMap runtime.Object
		errString string
		eventType watch.EventType
	}{
		{
			name: "invalid json",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "k8eraid-config"},
				Data:       map[string]string{"config.json": "{"},
			},
			errString: "unable to parse",
			eventType: watch.Added,
		},
		{
			name: "missing config",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "k8eraid-config"},
				Data:       map[string]string{},
			},
			errString: "missing config.json",
			eventType: watch.Added,
		},
		{
			name: "not a configmap",
			configMap: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "k8eraid-config"},
			},
			errString: "unable to coerce event",
			eventType: watch.Added,
		},
		{
			configMap: nil,
			eventType: watch.Deleted,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(subT *testing.T) {
			event := watch.Event{
				Type:   test.eventType,
				Object: test.configMap,
			}
			if err := eventReceived(event, &config); err == nil {
				subT.Error("eventReceived did not return an error")
			} else {
				if !strings.Contains(err.Error(), test.errString) {
					subT.Errorf("eventReceived returned %s, expected %s", err.Error(), test.errString)
				}
			}
		})
	}
}
