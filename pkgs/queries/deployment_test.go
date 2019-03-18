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

func Test_PollDeployment_ok(t *testing.T) {

	_, conf := StubsInit()

	tests := []struct {
		alertSpec      DeploymentAlertSpec
		name           string
		deployment     *appsv1.Deployment
		shouldAlert    bool
		alertersConfig AlertersConfig
	}{
		{
			name: "simple deployment, no alert",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deployment",
					Namespace: metav1.NamespaceDefault,
				},
			},
			alertSpec: DeploymentAlertSpec{
				Name:      "test-deployment",
				DepFilter: metav1.NamespaceDefault,
			},
			alertersConfig: conf,
		},
		{
			name: "simple deployment, missing replicas",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.Time{Time: time.Now().Add(time.Second * -10)},
					Name:              "test-deployment",
					Namespace:         metav1.NamespaceDefault,
				},
				Status: appsv1.DeploymentStatus{
					AvailableReplicas: 1,
				},
			},
			alertSpec: DeploymentAlertSpec{
				Name:      "test-deployment",
				DepFilter: metav1.NamespaceDefault,
				ReportStatus: DeploymentAlertStatus{
					PendingThreshold: 5,
					MinReplicas:      2,
				},
			},
			shouldAlert:    true,
			alertersConfig: conf,
		},
		{
			name: "wildcard, no alert",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deployment",
					Namespace: metav1.NamespaceDefault,
				},
			},
			alertSpec: DeploymentAlertSpec{
				Name: "*",
			},
			alertersConfig: conf,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(subT *testing.T) {
			client := fake.NewSimpleClientset(test.deployment)
			stubCalled := false
			alertStub := func(_ string, _ string, _ string, _ AlertersConfig) {
				stubCalled = true
			}
			err := PollDeployment(client, test.alertSpec, defaultTickerTime, alertStub, conf)
			if err != nil {
				subT.Errorf("PollDeployment returned an unexpected error: %s", err.Error())
				subT.Fail()
			}
			if test.shouldAlert != stubCalled {
				subT.Error("alert function should/should not have been called and was/was not")
				subT.Fail()
			}
		})
	}
}
