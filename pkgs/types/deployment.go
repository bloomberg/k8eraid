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

// DeploymentAlertStatus represents the thresholds to alert on for Deployments
type DeploymentAlertStatus struct {
	MinReplicas      int32 `json:"minReplicas"`
	PendingThreshold int64 `json:"pendingThreshold"`
}

// DeploymentAlertSpec represents a Deployment Alert Rule
type DeploymentAlertSpec struct {
	Name         string                `json:"name"`
	DepFilter    string                `json:"filter"`
	AlerterType  string                `json:"alerterType"`
	AlerterName  string                `json:"alerterName"`
	ReportStatus DeploymentAlertStatus `json:"reportStatus"`
}
