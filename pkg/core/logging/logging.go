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

package logging

import (
	"github.com/sirupsen/logrus"
)

var (
	gearModule string
)

func SetModule(module string) {
	gearModule = module
}

func LoggerEntry(name string) *logrus.Entry {
	logger := logrus.New()
	fields := logrus.Fields{}
	if name != "" {
		fields["logger"] = name
	}
	if gearModule != "" {
		fields["module"] = gearModule
	}
	return logger.WithFields(fields)
}
