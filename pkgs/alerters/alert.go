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

package alerters

import (
	"log"
	"os"

	"github.com/bloomberg/k8eraid/pkgs/types"
)

var (
	logger    *log.Logger
	errLogger *log.Logger
)

func init() {
	// Set up stdout and stderr loggers
	logger = log.New(os.Stdout, "alerters", log.LstdFlags)
	errLogger = log.New(os.Stderr, "alerters", log.LstdFlags)
}

// Alert function takes alertType, alertName and alertMessage as inputs, and triggers the correct alert type
func Alert(
	alertType string,
	alertName string,
	alertMessage string,
	config types.AlertersConfig,
) {

	// if alert type is stderr or blank, alert to stderr
	if alertType == "stderr" || alertType == "" {
		AlertStderr(alertMessage)
	}

	// if alert type is smtp, find matching rule and send mail
	if alertType == "smtp" {
		for _, alertRules := range config.Types.SMTPAlerterList {
			if alertRules.Name == alertName {
				AlertSMTP(alertRules, alertMessage)
			}
		}
	}

	// if alert type is pagerdutyV2, find matching rule and alert
	if alertType == "pagerdutyV2" {
		for _, alertRules := range config.Types.PDAlerterList {
			if alertRules.Name == alertName {
				AlertPagerDuty(alertRules, alertMessage)
			}
		}
	}

	// if alert type is webhook, find matching rule and trigger hook
	if alertType == "webhook" {
		for _, alertRules := range config.Types.WebhookAlerterList {
			if alertRules.Name == alertName {
				AlertWebhook(alertRules, alertMessage)
			}
		}
	}
	if alertType == "slack" {
		for _, alertRules := range config.Types.SlackAlerterList {
			if alertRules.Name == alertName {
				AlertSlack(alertRules, alertMessage)
			}
		}
	}
}
