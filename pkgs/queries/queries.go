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

package queries

import (
	"log"

	"github.com/bloomberg/k8eraid/pkgs/types"
)

var (
	logger    *log.Logger
	errLogger *log.Logger
	timeout   = int64(5)
)

// PollErr is an error returned by Poll*
type PollErr struct {
	Message string
}

func (err *PollErr) Error() string {
	return err.Message
}

type alertFunction func(string, string, string, types.AlertersConfig)
