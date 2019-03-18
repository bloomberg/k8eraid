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

package queries

import (
	"testing"
	"time"

	. "github.com/bloomberg/k8eraid/pkgs/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_PollNode_ok(t *testing.T) {

	_, conf := StubsInit()

	tests := []struct {
		alertSpec      NodeAlertSpec
		name           string
		node           *corev1.Node
		shouldAlert    bool
		alertersConfig AlertersConfig
	}{
		{
			name: "basic node: no alert",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-node",
					Namespace: metav1.NamespaceDefault,
				},
			},
			alertSpec:      NodeAlertSpec{},
			alertersConfig: conf,
		},
		{
			name: "wildcard, basic node, less than min nodes: alert",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-node",
					Namespace: metav1.NamespaceDefault,
				},
			},
			alertSpec: NodeAlertSpec{
				Name: "*",
				ReportStatus: NodeAlertStatus{
					MinNodes: 2,
				},
			},
			shouldAlert:    true,
			alertersConfig: conf,
		},
		{
			name: "basic node with conditions, ready: alert",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.Time{Time: time.Now().Add(time.Second * -10)},
					Name:              "test-node",
					Namespace:         metav1.NamespaceDefault,
				},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							LastTransitionTime: metav1.Time{Time: time.Now().Add(time.Second * -40)},
							Type:               corev1.NodeReady,
						},
					},
				},
			},
			alertSpec: NodeAlertSpec{
				ReportStatus: NodeAlertStatus{
					PendingThreshold: 5,
					NodeReady:        true,
				},
			},
			shouldAlert:    true,
			alertersConfig: conf,
		},
		{
			name: "basic node with conditions, OutOfDisk: alert",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.Time{Time: time.Now().Add(time.Second * -10)},
					Name:              "test-node",
					Namespace:         metav1.NamespaceDefault,
				},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							LastTransitionTime: metav1.Time{Time: time.Now().Add(time.Second * -40)},
							Type:               corev1.NodeOutOfDisk,
						},
					},
				},
			},
			alertSpec: NodeAlertSpec{
				ReportStatus: NodeAlertStatus{
					PendingThreshold: 5,
					NodeOutOfDisk:    true,
				},
			},
			shouldAlert:    true,
			alertersConfig: conf,
		},
		{
			name: "basic node with conditions, MemoryPressure: alert",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.Time{Time: time.Now().Add(time.Second * -10)},
					Name:              "test-node",
					Namespace:         metav1.NamespaceDefault,
				},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							LastTransitionTime: metav1.Time{Time: time.Now().Add(time.Second * -40)},
							Type:               corev1.NodeMemoryPressure,
						},
					},
				},
			},
			alertSpec: NodeAlertSpec{
				ReportStatus: NodeAlertStatus{
					PendingThreshold:   5,
					NodeMemoryPressure: true,
				},
			},
			shouldAlert:    true,
			alertersConfig: conf,
		},
		{
			name: "basic node with conditions, DiskPressure: alert",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.Time{Time: time.Now().Add(time.Second * -10)},
					Name:              "test-node",
					Namespace:         metav1.NamespaceDefault,
				},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							LastTransitionTime: metav1.Time{Time: time.Now().Add(time.Second * -40)},
							Type:               corev1.NodeDiskPressure,
						},
					},
				},
			},
			alertSpec: NodeAlertSpec{
				ReportStatus: NodeAlertStatus{
					PendingThreshold: 5,
					NodeDiskPressure: true,
				},
			},
			shouldAlert:    true,
			alertersConfig: conf,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(subT *testing.T) {
			client := fake.NewSimpleClientset(test.node)
			stubCalled := false
			alertStub := func(_ string, _ string, _ string, _ AlertersConfig) {
				stubCalled = true
			}
			err := PollNode(client, test.alertSpec, defaultTickerTime, alertStub, conf)
			if err != nil {
				subT.Errorf("PollNode returned an unexpected error: %s", err.Error())
				subT.Fail()
			}
			if test.shouldAlert != stubCalled {
				subT.Error("alert function should have been called and was not")
				subT.Fail()
			}
		})
	}
}
