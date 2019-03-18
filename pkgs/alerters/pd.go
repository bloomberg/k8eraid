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
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	pagerduty "github.com/PagerDuty/go-pagerduty"
	"github.com/bloomberg/k8eraid/pkgs/types"
)

// AlertPagerDuty triggers Pager Duty alerts via the v2API using data relayed from alerts.go
func AlertPagerDuty(alertdata types.PDAlerterConfig, message string) {
	myEvent, myClient := PagerDutyInput(alertdata, message)
	resp := PagerDutyTrigger(myEvent, myClient)
	logger.Print(resp)
}

// PagerDutyInput generates the formatted alert inputs for triggering a pagerduty alert
func PagerDutyInput(a types.PDAlerterConfig, m string) (pagerduty.Event, *http.Client) {
	// Get key from ENV that was specified
	keyenvvar := a.ServiceKeyEnvVar
	key := os.Getenv(keyenvvar)
	mytime := time.Now().Local()

	// Specify alert details
	D := types.PDAlertDetails{
		Subject: a.Subject,
		Message: m,
		Time:    mytime,
	}

	// Construct event
	event := pagerduty.Event{
		Type:        "trigger",
		ServiceKey:  key,
		Description: a.Subject,
		Details:     D,
	}

	var myTransport *http.Transport
	const (
		timeout5  = 5 * time.Second
		timeout10 = 10 * time.Second
	)

	// Set http proxy and custom http client
	if a.ProxyServer != "" {
		proxyURL, _ := url.Parse(a.ProxyServer)
		myTransport = &http.Transport{
			Dial: (&net.Dialer{
				Timeout: timeout5,
			}).Dial,
			TLSHandshakeTimeout: timeout5,
			Proxy:               http.ProxyURL(proxyURL),
		}
	} else {
		myTransport = &http.Transport{
			Dial: (&net.Dialer{
				Timeout: timeout5,
			}).Dial,
			TLSHandshakeTimeout: timeout5,
		}
	}

	var myClient = &http.Client{
		Timeout:   timeout10,
		Transport: myTransport,
	}
	return event, myClient
}

// PagerDutyTrigger triggers a pagerduty alert
func PagerDutyTrigger(e pagerduty.Event, c *http.Client) string {
	var err error
	resp, err := pagerduty.CreateEventWithHTTPClient(e, c)
	if err != nil {
		return "Issue sending PagerDuty alert: " + err.Error()
	}
	return "Pager Duty incident triggered, key: " + resp.IncidentKey
}
