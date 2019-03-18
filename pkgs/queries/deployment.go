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
	"fmt"
	"strings"
	"time"

	"github.com/bloomberg/k8eraid/pkgs/types"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PollDeployment function takes inputs and iterates across deployments in the kubernetes cluster, triggering alerts as needed.
func PollDeployment(
	clientset kubernetes.Interface,
	alertSpec types.DeploymentAlertSpec,
	tickertime int64,
	alertFn alertFunction,
	alertersConfig types.AlertersConfig,
) error {

	if alertSpec.ReportStatus.PendingThreshold == 0 {
		alertSpec.ReportStatus.PendingThreshold = 10
	}

	// If the deployment is not wildcard, search by name
	if alertSpec.Name != "*" {
		if alertSpec.DepFilter == "" {
			return &PollErr{
				Message: fmt.Sprintf("Deployment rule for %s has no namespace filter specified, ignoring\n", alertSpec.Name),
			}
		}

		// Get the deployment
		deployment, deploymenterr := clientset.AppsV1().Deployments(alertSpec.DepFilter).Get(alertSpec.Name, metav1.GetOptions{})
		if deploymenterr != nil {
			return &PollErr{
				Message: fmt.Sprintf("Error fetching deployment: %s", deploymenterr.Error()),
			}
		}
		checkDeployment(deployment, alertSpec, alertFn, alertersConfig)

		// If the deployment is a wildcard, list deployments and iterate through
	} else {
		listopts := metav1.ListOptions{
			LabelSelector:        alertSpec.DepFilter,
			IncludeUninitialized: false,
			Watch:                false,
			TimeoutSeconds:       &timeout,
		}
		if strings.Contains(alertSpec.DepFilter, "=") || alertSpec.DepFilter == "" {
			deployments, deploymentserr := clientset.AppsV1().Deployments("").List(listopts)
			if deploymentserr != nil {
				return &PollErr{
					Message: fmt.Sprintf("Unable to get deployments: %s", deploymentserr.Error()),
				}
			}
			for _, deploymentname := range deployments.Items {
				deployment, deploymenterr := clientset.AppsV1().Deployments(deploymentname.GetNamespace()).Get(deploymentname.GetName(), metav1.GetOptions{})
				if deploymenterr != nil {
					return &PollErr{
						Message: fmt.Sprintf("Error fetching deployment %s: %s", deploymentname.GetName(), deploymenterr.Error()),
					}
				}
				checkDeployment(deployment, alertSpec, alertFn, alertersConfig)
			}
		} else {

			return &PollErr{
				Message: fmt.Sprint("Deployment rule for global has incorrect filter specified, ignoring"),
			}
		}
	}
	return nil
}

func checkDeployment(
	deployment *appsv1.Deployment,
	alertSpec types.DeploymentAlertSpec,
	alertFn alertFunction,
	alertersConfig types.AlertersConfig,
) {

	// Get times for comparing to threshold
	statusCreatedSecondsDiff := time.Now().Unix() - deployment.ObjectMeta.CreationTimestamp.Unix()

	// If deployment hasnt been around longer than threshold, bail. otherwise check the status.
	if statusCreatedSecondsDiff > alertSpec.ReportStatus.PendingThreshold {
		if deployment.Status.AvailableReplicas < alertSpec.ReportStatus.MinReplicas {
			// ALERT
			s := []string{"Deployment", alertSpec.Name, "does not have the specified required minimum replicas"}
			alertmessage := strings.Join(s, " ")
			alertFn(alertSpec.AlerterType, alertSpec.AlerterName, alertmessage, alertersConfig)
		}
	}
}
