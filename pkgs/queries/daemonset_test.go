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

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_PollDaemonset_ok(t *testing.T) {

	_, conf := StubsInit()

	tests := []struct {
		alertSpec      DaemonsetAlertSpec
		name           string
		daemonSet      *appsv1.DaemonSet
		shouldAlert    bool
		alertersConfig AlertersConfig
	}{
		{
			name: "basic daemonset, no alert",
			daemonSet: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-daemonset",
					Namespace: metav1.NamespaceDefault,
				},
			},
			alertSpec: DaemonsetAlertSpec{
				Name:         "test-daemonset",
				DaemonFilter: metav1.NamespaceDefault,
			},
			shouldAlert:    false,
			alertersConfig: conf,
		},
		{
			name: "basic daemonset, recently created and not available, alert",
			daemonSet: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.Time{Time: time.Now().Add(time.Second * -10)},
					Name:              "test-daemonset",
					Namespace:         metav1.NamespaceDefault,
				},
				Status: appsv1.DaemonSetStatus{
					CurrentNumberScheduled: 0,
					NumberAvailable:        1,
				},
			},
			alertSpec: DaemonsetAlertSpec{
				Name:         "test-daemonset",
				DaemonFilter: metav1.NamespaceDefault,
				ReportStatus: DaemonsetAlertStatus{
					CheckReplicas:    true,
					PendingThreshold: 5,
				},
			},
			shouldAlert:    true,
			alertersConfig: TestAlertersConfig,
		},
		{
			name: "basic daemonset, failed scheduling, alert",
			daemonSet: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-daemonset",
					Namespace: metav1.NamespaceDefault,
				},
				Status: appsv1.DaemonSetStatus{
					CurrentNumberScheduled: 1,
					DesiredNumberScheduled: 2,
				},
			},
			alertSpec: DaemonsetAlertSpec{
				Name:         "test-daemonset",
				DaemonFilter: metav1.NamespaceDefault,
				ReportStatus: DaemonsetAlertStatus{
					FailedScheduling: true,
				},
			},
			shouldAlert:    true,
			alertersConfig: conf,
		},
		{
			name: "wildcard, basic daemonset, no alert",
			daemonSet: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-daemonset",
					Namespace: metav1.NamespaceDefault,
					Labels: map[string]string{
						"foo": "bar",
					},
				},
			},
			alertSpec: DaemonsetAlertSpec{
				Name:         "*",
				DaemonFilter: "foo=bar",
			},
			shouldAlert:    false,
			alertersConfig: conf,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(subT *testing.T) {
			client := fake.NewSimpleClientset(test.daemonSet)
			stubCalled := false
			alertStub := func(_ string, _ string, _ string, _ AlertersConfig) {
				stubCalled = true
			}
			err := PollDaemonset(client, test.alertSpec, defaultTickerTime, alertStub, conf)
			if err != nil {
				subT.Errorf("PollDaemonset returned an unexpected error: %s", err.Error())
				subT.Fail()
			}
			if test.shouldAlert != stubCalled {
				subT.Error("alert function should have been called and was not")
				subT.Fail()
			}
		})
	}
}
