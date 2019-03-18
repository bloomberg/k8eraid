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

var (
	defaultMeta = metav1.ObjectMeta{
		Name:      "test-pod",
		Namespace: metav1.NamespaceDefault,
	}
	defaultDeletionGracePeriod = int64(5)
)

func Test_PullPod_ok(t *testing.T) {

	_, conf := StubsInit()

	tests := []struct {
		alertSpec      PodAlertSpec
		name           string
		pod            *corev1.Pod
		shouldAlert    bool
		alertersConfig AlertersConfig
	}{
		{
			name: "basic pod: no alert",
			pod: &corev1.Pod{
				ObjectMeta: defaultMeta,
			},
			alertSpec: PodAlertSpec{
				Name:               "test-pod",
				PodFilterNamespace: metav1.NamespaceDefault,
			},
			shouldAlert:    false,
			alertersConfig: TestAlertersConfig,
		},
		{
			name: "pod with conditions: no alert",
			pod: &corev1.Pod{
				ObjectMeta: defaultMeta,
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:               corev1.PodReady,
							LastTransitionTime: metav1.Time{Time: time.Now().Add(time.Duration(-defaultTickerTime))},
						},
					},
				},
			},
			alertSpec: PodAlertSpec{
				Name:               "test-pod",
				PodFilterNamespace: metav1.NamespaceDefault,
			},
			shouldAlert:    false,
			alertersConfig: conf,
		},
		{
			name: "pod with conditions, ready, restarted: alert",
			pod: &corev1.Pod{
				ObjectMeta: defaultMeta,
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:               corev1.PodReady,
							LastTransitionTime: metav1.Time{Time: time.Now().Add(time.Duration(-(defaultTickerTime + 5)))},
						},
					},
				},
			},
			alertSpec: PodAlertSpec{
				Name:               "test-pod",
				PodFilterNamespace: metav1.NamespaceDefault,
				ReportStatus: PodAlertStatus{
					PodRestarts: true,
				},
			},
			shouldAlert:    true,
			alertersConfig: conf,
		},
		{
			name: "pod with conditions, scheduled, failed to schedule: alert",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: metav1.NamespaceDefault,
				},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodScheduled,
							Status: corev1.ConditionFalse,
						},
					},
				},
			},
			alertSpec: PodAlertSpec{
				Name:               "test-pod",
				PodFilterNamespace: metav1.NamespaceDefault,
				ReportStatus: PodAlertStatus{
					FailedScheduling: true,
				},
			},
			shouldAlert:    true,
			alertersConfig: conf,
		},
		{
			name: "pod stuck terminating: alert",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp:          &metav1.Time{Time: time.Now().Add(time.Second * -defaultTickerTime)},
					DeletionGracePeriodSeconds: &defaultDeletionGracePeriod,
					Name:                       "test-pod",
					Namespace:                  metav1.NamespaceDefault,
				},
			},
			alertSpec: PodAlertSpec{
				PodFilterNamespace: metav1.NamespaceDefault,
				ReportStatus: PodAlertStatus{
					StuckTerminating: true,
				},
			},
			shouldAlert:    true,
			alertersConfig: TestAlertersConfig,
		},
		{
			name: "wildcard, no alert",
			pod: &corev1.Pod{
				ObjectMeta: defaultMeta,
			},
			alertSpec: PodAlertSpec{
				Name: "*",
			},
			shouldAlert:    false,
			alertersConfig: conf,
		},
		{
			name: "wildcard, no matching pods, alert",
			pod:  &corev1.Pod{},
			alertSpec: PodAlertSpec{
				Name: "*",
				ReportStatus: PodAlertStatus{
					MinPods: 1,
				},
				PodFilterLabel: "foo=bar",
			},
			shouldAlert:    true,
			alertersConfig: conf,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(subT *testing.T) {
			client := fake.NewSimpleClientset(test.pod)
			stubCalled := false
			alertStub := func(_ string, _ string, _ string, _ AlertersConfig) {
				stubCalled = true
			}
			err := PollPod(client, test.alertSpec, defaultTickerTime, alertStub, conf)
			if err != nil {
				subT.Errorf("PollPod returned an unexpected error: %s", err.Error())
				subT.Fail()
			}
			if test.shouldAlert != stubCalled {
				subT.Error("alert function should have been called and was not")
				subT.Fail()
			}
		})
	}
}
