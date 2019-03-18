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
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/bloomberg/k8eraid/pkgs/types"

	"github.com/nlopes/slack"
)

// AlertSlack sends an alert to slack
func AlertSlack(alertData types.SlackAlerterConfig, message string) {
	msg := SlackInput(message)
	origTransport := http.DefaultTransport
	if alertData.ProxyServer != "" {
		// we need to override the default transport to apply proxy settings
		proxyURL, err := url.Parse(alertData.ProxyServer)
		if err != nil {
			log.Printf(
				"Alert configuration %s has invalid proxy server(%s): %s",
				alertData.Name,
				alertData.ProxyServer,
				err.Error(),
			)
			return
		}
		http.DefaultTransport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 10 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       10 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}
	if err := slack.PostWebhook(alertData.WebhookURL, msg); err != nil {
		log.Printf("Error sending alert to Slack: %s", err.Error())
	}
	http.DefaultTransport = origTransport
}

// SlackInput formats an alert for Slack
func SlackInput(message string) *slack.WebhookMessage {
	attach := slack.Attachment{
		Fallback:   message,
		Color:      "#ff0000",
		AuthorName: "k8eraid",
		AuthorIcon: "https://github.com/kubernetes/kubernetes/raw/master/logo/logo.png",
		Title:      "k8eraid alert",
		Text:       message,
		Ts:         json.Number(fmt.Sprint(time.Now().Unix())),
	}
	return &slack.WebhookMessage{Attachments: []slack.Attachment{attach}}
}
