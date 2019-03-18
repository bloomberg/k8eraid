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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bloomberg/k8eraid/pkgs/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const expected = `{"attachments":[{"color":"#ff0000","fallback":"foo","author_name":"k8eraid","author_icon":"https://github.com/kubernetes/kubernetes/raw/master/logo/logo.png","title":"k8eraid alert","text":"foo","ts":%d}]}`

func Test_AlertSlack_OK(t *testing.T) {
	withWebhookServer(t, false, func(buf *bytes.Buffer, url string) {
		AlertSlack(types.SlackAlerterConfig{WebhookURL: url}, "foo")
		assert.Equal(t, fmt.Sprintf(expected, time.Now().Unix()), string(buf.Bytes()), "Expected request data should match actual")
	})
}

func withWebhookServer(t *testing.T, fail bool, f func(buf *bytes.Buffer, url string)) {
	buf := &bytes.Buffer{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_, err := io.Copy(buf, r.Body)
		require.NoError(t, err, "io.Copy should not return an error")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	f(buf, server.URL)
}
