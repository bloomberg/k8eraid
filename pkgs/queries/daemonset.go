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

// PollDaemonset function takes inputs and iterates across daemonsets in the kubernetes cluster, triggering alerts as needed.
func PollDaemonset(
	clientset kubernetes.Interface,
	alertSpec types.DaemonsetAlertSpec,
	tickertime int64,
	alertFn alertFunction,
	alertersConfig types.AlertersConfig,
) error {

	if alertSpec.ReportStatus.PendingThreshold == 0 {
		alertSpec.ReportStatus.PendingThreshold = 10
	}

	// If the daemon is not wildcard, search by name
	if alertSpec.Name != "*" {
		if alertSpec.DaemonFilter == "" {
			return &PollErr{
				Message: fmt.Sprintf("Daemonset rule for %s has no namespace filter specified, ignoring", alertSpec.Name),
			}
		}
		daemonset, daemonseterr := clientset.AppsV1().DaemonSets(alertSpec.DaemonFilter).Get(alertSpec.Name, metav1.GetOptions{})
		if daemonseterr != nil {
			return &PollErr{
				Message: fmt.Sprintf("Error fetching daemonset %s: %s", alertSpec.Name, daemonseterr.Error()),
			}
		}

		checkDaemonset(daemonset, alertSpec, alertFn, alertersConfig)
		// If the daemon is a wildcard, list daemons and iterate through
	} else {
		if strings.Contains(alertSpec.DaemonFilter, "=") || alertSpec.DaemonFilter == "" {
			listopts := metav1.ListOptions{
				LabelSelector:        alertSpec.DaemonFilter,
				IncludeUninitialized: false,
				Watch:                false,
				TimeoutSeconds:       &timeout,
			}
			daemonsets, daemonsetserr := clientset.AppsV1().DaemonSets("").List(listopts)
			if daemonsetserr != nil {
				return &PollErr{
					Message: fmt.Sprintf("Unable to list DaemonSets: %s", daemonsetserr.Error()),
				}
			}
			for _, daemonsetname := range daemonsets.Items {
				daemonset, daemonseterr := clientset.AppsV1().DaemonSets(daemonsetname.GetNamespace()).Get(daemonsetname.GetName(), metav1.GetOptions{})
				if daemonseterr != nil {
					return &PollErr{
						Message: fmt.Sprintf("Unable to get DaemonSet %s: %s", daemonsetname.Name, daemonsetserr.Error()),
					}
				}
				checkDaemonset(daemonset, alertSpec, alertFn, alertersConfig)
			}
		} else {
			return &PollErr{
				Message: fmt.Sprintf("Deployment rule for global has incorrect filter specified (filter was: %s), ignoring", alertSpec.DaemonFilter),
			}
		}
	}
	return nil
}

func checkDaemonset(
	daemonSet *appsv1.DaemonSet,
	alertSpec types.DaemonsetAlertSpec,
	alertFn alertFunction,
	alertersConfig types.AlertersConfig,
) {
	nowSeconds := time.Now().Unix()
	// Get times for comparing to threshold
	statusCreatedSecondsDiff := nowSeconds - daemonSet.ObjectMeta.CreationTimestamp.Unix()

	// If daemonset hasnt been around longer than threshold, bail. otherwise check the status.
	if statusCreatedSecondsDiff > alertSpec.ReportStatus.PendingThreshold {
		statusReplicas := daemonSet.Status.CurrentNumberScheduled
		if alertSpec.ReportStatus.CheckReplicas {
			if statusReplicas < daemonSet.Status.NumberAvailable {
				// ALERT
				alertmessage := fmt.Sprint(
					"Daemonset",
					alertSpec.Name,
					"in namespace",
					alertSpec.DaemonFilter,
					"does not have the specified required minimum replicas available!",
				)
				alertFn(alertSpec.AlerterType, alertSpec.AlerterName, alertmessage, alertersConfig)
			}
		}
		if alertSpec.ReportStatus.FailedScheduling {
			if statusReplicas < daemonSet.Status.DesiredNumberScheduled {
				// ALERT
				alertmessage := fmt.Sprint(
					"Daemonset",
					alertSpec.Name,
					"in namespace",
					alertSpec.DaemonFilter,
					"does not have the desired number of replicas scheduled!",
				)
				alertFn(alertSpec.AlerterType, alertSpec.AlerterName, alertmessage, alertersConfig)
			}
		}
	}
}
