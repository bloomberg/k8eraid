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
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/bloomberg/k8eraid/pkgs/types"
)

// AlertWebhook sends a general http(s) payload using data relayed from alerts.go
func AlertWebhook(alertdata types.WebhookAlerterConfig, message string) {
	mytime := time.Now().Local()

	// Specify alert details
	D := types.WebhookAlertDetails{}
	D.Subject = alertdata.Subject
	D.Msg = message
	D.Time = mytime

	// Set http proxy and custom http client

	var myTransport *http.Transport
	var err error

	if alertdata.ProxyServer != "" {
		proxyURL, _ := url.Parse(alertdata.ProxyServer)
		myTransport = &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
			Proxy:               http.ProxyURL(proxyURL),
		}
	} else {
		myTransport = &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		}
	}

	var myClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: myTransport,
	}

	// Trigger event
	err = createWebhookWithHTTPClient(D, myClient, alertdata)
	if err != nil {
		errLogger.Println("Issue sending Webhook alert: ", err)
	}
	logger.Println("Webhook event triggered for webhook alerter: ", alertdata.Name)
}

func createWebhookWithHTTPClient(d types.WebhookAlertDetails, client *http.Client, alertdata types.WebhookAlerterConfig) error {
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}
	req, _ := http.NewRequest("POST", alertdata.Server, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP Status Code: %d", resp.StatusCode)
	}
	return nil
}
