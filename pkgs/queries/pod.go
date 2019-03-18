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
	"time"

	"github.com/bloomberg/k8eraid/pkgs/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PollPod function takes inputs and iterates across pods in the kubernetes cluster, triggering alerts as needed.
func PollPod(
	clientset kubernetes.Interface,
	alertSpec types.PodAlertSpec,
	tickertime int64,
	alertFn alertFunction,
	alertersConfig types.AlertersConfig,
) error {

	if alertSpec.ReportStatus.PendingThreshold == 0 {
		alertSpec.ReportStatus.PendingThreshold = 10
	}

	// Check rules with matching literal pod name
	if alertSpec.Name != "*" {
		if alertSpec.PodFilterNamespace == "" {
			return &PollErr{
				Message: fmt.Sprintf("pod rule for %s has no namespace filter specified, ignoring\n", alertSpec.Name),
			}
		}

		pod, poderr := clientset.CoreV1().Pods(alertSpec.PodFilterNamespace).Get(alertSpec.Name, metav1.GetOptions{})
		if poderr != nil {
			return &PollErr{
				Message: fmt.Sprintf("error getting pod %s: %s", alertSpec.Name, poderr.Error()),
			}
		}
		checkPod(pod, alertSpec, tickertime, alertFn, alertersConfig)
		// If podname is a wildcard, list based on filter and iterate through
	} else {
		listopts := metav1.ListOptions{
			LabelSelector:        alertSpec.PodFilterLabel,
			IncludeUninitialized: false,
			Watch:                false,
			TimeoutSeconds:       &timeout,
		}
		// Check rules by label
		pods, podserr := clientset.CoreV1().Pods("").List(listopts)
		if podserr != nil {
			return &PollErr{
				Message: fmt.Sprintf("error fetching pods: %s", podserr.Error()),
			}
		}

		// Check to see if there are the minimum specified pods matching rule
		if len(pods.Items) < int(alertSpec.ReportStatus.MinPods) {
			// ALERT
			alertmessage := fmt.Sprint("Number of pods for label", alertSpec.PodFilterLabel, "is under minimum specification!")
			alertFn(alertSpec.AlerterType, alertSpec.AlerterName, alertmessage, alertersConfig)
		}

		// Iterate through pod items
		for _, poddata := range pods.Items {
			pod, poderr := clientset.CoreV1().Pods(poddata.GetNamespace()).Get(poddata.GetName(), metav1.GetOptions{})
			if poderr != nil {
				return &PollErr{
					Message: fmt.Sprintf("Unable to get pod %s: %s", poddata.Name, poderr.Error()),
				}
			}

			checkPod(pod, alertSpec, tickertime, alertFn, alertersConfig)
		}
	}
	return nil
}

func checkPod(
	pod *corev1.Pod,
	alertSpec types.PodAlertSpec,
	tickertime int64,
	alertFn alertFunction,
	alertersConfig types.AlertersConfig,
) {
	nowSeconds := time.Now().Unix()
	// Get times for comparing to threshold
	statusCreatedSecondsDiff := nowSeconds - pod.ObjectMeta.CreationTimestamp.Unix()

	// If pod hasnt been around longer than threshold, bail. otherwise check the status.
	if statusCreatedSecondsDiff > alertSpec.ReportStatus.PendingThreshold {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == "Ready" {
				transitiontimeDiff := time.Now().Unix() - condition.LastTransitionTime.Unix()
				if transitiontimeDiff < tickertime && alertSpec.ReportStatus.PodRestarts {
					// ALERT
					alertmessage := fmt.Sprint("Pod", alertSpec.Name, "has changed ready status since last poll and may be restarting!")
					alertFn(alertSpec.AlerterType, alertSpec.AlerterName, alertmessage, alertersConfig)
				}
			} else if condition.Type == "PodScheduled" {
				if condition.Status != "True" && alertSpec.ReportStatus.FailedScheduling {
					// ALERT
					alertmessage := fmt.Sprint("Pod", alertSpec.Name, "has not been scheduled yet and has passed scheduling timeline!")
					alertFn(alertSpec.AlerterType, alertSpec.AlerterName, alertmessage, alertersConfig)
				}
			}
		}
	}

	// Check for stuck in terminating status.
	if pod.ObjectMeta.DeletionTimestamp != nil && alertSpec.ReportStatus.StuckTerminating == true {
		deletionGracePeriod := *pod.ObjectMeta.DeletionGracePeriodSeconds

		// delete scheduled + grace period
		deletionDeadline := pod.ObjectMeta.DeletionTimestamp.Unix() + deletionGracePeriod

		// Get the time of the last status check
		lastpollDiff := nowSeconds - tickertime

		// If the deletion deadline has passed within the time of last status check, alert
		if deletionDeadline < nowSeconds && deletionDeadline > lastpollDiff {
			// ALERT
			alertmessage := fmt.Sprint("Pod", alertSpec.Name, "has passed its deletion timeline and may be stuck in terminating status!")
			alertFn(alertSpec.AlerterType, alertSpec.AlerterName, alertmessage, alertersConfig)
		}
	}
}
