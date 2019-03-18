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
	"net/smtp"
	"os"
	"strconv"

	"github.com/bloomberg/k8eraid/pkgs/types"
)

// AlertSMTP send SMTP messages using inputs forwarded from alert.go
func AlertSMTP(alertdata types.SMTPAlerterConfig, message string) {
	from := alertdata.FromAddress
	to := alertdata.ToAddress
	subject := alertdata.Subject
	pass := os.Getenv(alertdata.PasswordEnvVar)

	port := strconv.Itoa(alertdata.Port)
	server := alertdata.MailServer + ":" + port

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		message

	err := smtp.SendMail(server,
		smtp.PlainAuth("", from, pass, alertdata.MailServer),
		from, []string{to}, []byte(msg))

	if err != nil {
		errLogger.Print("smtp error: ", err)
	} else {
		logger.Print("Alert message sent to ", to)
	}
}
