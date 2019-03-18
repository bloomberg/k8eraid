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

var (
	// TestConfigRules holds the testable ConfigRules{} struct
	TestConfigRules ConfigRules
	// TestAlertersConfig holds the testable AlertersConfig{} struct
	TestAlertersConfig AlertersConfig
)

// StubsInit generates stub configurations purely for testing purposes
func StubsInit() (ConfigRules, AlertersConfig) {
	TestConfigRules = ConfigRules{
		Deployments: []DeploymentAlertSpec{
			{
				Name:        "kube-dns",
				DepFilter:   "kube-system",
				AlerterType: "stderr",
				ReportStatus: DeploymentAlertStatus{
					MinReplicas:      1,
					PendingThreshold: 1,
				},
			},
		},
		Pods: []PodAlertSpec{
			{
				Name:               "*",
				PodFilterNamespace: "kube-system",
				PodFilterLabel:     "",
				AlerterType:        "smtp",
				AlerterName:        "example-email",
				ReportStatus: PodAlertStatus{
					MinPods:          1,
					PodRestarts:      true,
					FailedScheduling: true,
					StuckTerminating: true,
					PendingThreshold: 1,
				},
			},
		},
		Daemonsets: []DaemonsetAlertSpec{
			{
				Name:         "nginx-ingress-controller",
				DaemonFilter: "kube-system",
				AlerterType:  "smtp",
				AlerterName:  "example-email",
				ReportStatus: DaemonsetAlertStatus{
					CheckReplicas:    true,
					FailedScheduling: true,
					PendingThreshold: 1,
				},
			},
		},
		Nodes: []NodeAlertSpec{
			{
				Name:        "*",
				NodeFilter:  "",
				AlerterType: "smtp",
				AlerterName: "example-email",
				ReportStatus: NodeAlertStatus{
					PendingThreshold:   60,
					MinNodes:           3,
					NodeOutOfDisk:      true,
					NodeMemoryPressure: true,
					NodeDiskPressure:   true,
					NodeReady:          true,
				},
			},
		},
	}
	TestAlertersConfig = AlertersConfig{
		AlerterTypes{
			SMTPAlerterList: []SMTPAlerterConfig{
				{
					Name:        "example-email",
					ToAddress:   "kubernetes@example.net",
					FromAddress: "kubernetes@example.net",
					MailServer:  "smtp.example.com",
					Port:        25,
					Subject:     "Test smtp alerter alert from k8eraid",
				},
			},
			WebhookAlerterList: []WebhookAlerterConfig{
				{
					Name:        "example-webhook",
					Server:      "http://www.example.com",
					Subject:     "Test webhook alerter alert from k8eraid",
					ProxyServer: "",
				},
			},
			PDAlerterList: []PDAlerterConfig{
				{
					Name:             "example-pagerduty",
					ServiceKeyEnvVar: "MYPDKEY",
					ProxyServer:      "http://someproxy.example.com",
					Subject:          "Test Pagerduty Alerter alert from k8eraid",
				},
			},
		},
	}
	return TestConfigRules, TestAlertersConfig
}
