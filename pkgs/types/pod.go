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

package types

// PodAlertStatus represents the thresholds for alerting on Pods
type PodAlertStatus struct {
	MinPods          int32 `json:"minPods"`
	PodRestarts      bool  `json:"podRestarts"`
	FailedScheduling bool  `json:"failedScheduling"`
	PendingThreshold int64 `json:"pendingThreshold"`
	StuckTerminating bool  `json:"stuckTerminating"`
}

// PodAlertSpec represents the configuration for alerting on Pods
type PodAlertSpec struct {
	Name               string         `json:"name"`
	PodFilterNamespace string         `json:"filterNamespace"`
	PodFilterLabel     string         `json:"filterLabel"`
	AlerterType        string         `json:"alerterType"`
	AlerterName        string         `json:"alerterName"`
	ReportStatus       PodAlertStatus `json:"reportStatus"`
}
