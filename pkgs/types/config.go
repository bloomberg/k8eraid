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

import (
	"time"
)

// ConfigRules represents the structure of the config file for k8eraid
type ConfigRules struct {
	Deployments    []DeploymentAlertSpec `json:"deployments"`
	Pods           []PodAlertSpec        `json:"pods"`
	Daemonsets     []DaemonsetAlertSpec  `json:"daemonsets"`
	Nodes          []NodeAlertSpec       `json:"nodes"`
	AlertersConfig AlertersConfig        `json:"alerters"`
}

// Alerter types

// SMTPAlerterConfig struct contains the data needed to trigger an SMTP alert
type SMTPAlerterConfig struct {
	Name           string `json:"name"`
	ToAddress      string `json:"toAddress"`
	FromAddress    string `json:"fromAddress"`
	MailServer     string `json:"mailServer"`
	Port           int    `json:"port"`
	Subject        string `json:"subject"`
	PasswordEnvVar string `json:"passwordEnvVar"`
}

// PDAlerterConfig struct contains the needed data for triggering a Pager Duty type alert
type PDAlerterConfig struct {
	Name             string `json:"name"`
	ServiceKeyEnvVar string `json:"serviceKeyEnvVar"`
	ProxyServer      string `json:"proxyServer"`
	Subject          string `json:"subject"`
}

// PDAlertDetails contains the needed data to put into the body of a Pager Duty type alert
type PDAlertDetails struct {
	Subject string    `json:"subject"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

// WebhookAlerterConfig struct contains the data needed to trigger an SMTP alert
type WebhookAlerterConfig struct {
	Name        string `json:"name"`
	Server      string `json:"server"`
	ProxyServer string `json:"proxyServer"`
	Subject     string `json:"subject"`
}

// WebhookAlertDetails contains the needed data to put into the body of a Webhook type alert
type WebhookAlertDetails struct {
	Subject string    `json:"subject"`
	Msg     string    `json:"message"`
	Time    time.Time `json:"time"`
}

// AlerterTypes are the actual types of alerter structs
type AlerterTypes struct {
	PDAlerterList      []PDAlerterConfig      `json:"pagerdutyV2"`
	SlackAlerterList   []SlackAlerterConfig   `json:"slack"`
	SMTPAlerterList    []SMTPAlerterConfig    `json:"smtp"`
	WebhookAlerterList []WebhookAlerterConfig `json:"webhook"`
}

// AlertersConfig is the top level struct containing alerter configuration data
type AlertersConfig struct {
	Types AlerterTypes `json:"alerters"`
}

//SlackAlerterConfig configures a Slack Alerter
type SlackAlerterConfig struct {
	Name        string `json:"name"`
	WebhookURL  string `json:"webhookURL"`
	ProxyServer string `json:"proxyServer"`
}
