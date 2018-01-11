// Copyright 2015-present Oursky Ltd.
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

package skydb

import (
	"github.com/skygeario/skygear-server/pkg/server/logging"
	"time"
)

var log = logging.LoggerEntry("skydb")
var timeNow = func() time.Time { return time.Now().UTC() }

func MockTimeNowForTestingOnly(f func() time.Time) func() {
	timeNow = f
	return restoreMockTimeNowForTestingOnly
}

func restoreMockTimeNowForTestingOnly() {
	timeNow = func() time.Time { return time.Now().UTC() }
}
