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
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	loggers                sync.Map
	configureLoggerHandler func(string, *logrus.Logger)
	gearModule             string
)

func SetConfigureLoggerHandler(handler func(string, *logrus.Logger)) {
	configureLoggerHandler = handler
}

func SetModule(module string) {
	gearModule = module
}

func getLogger(name string) *logrus.Logger {
	l, ok := loggers.Load(name)
	var logger *logrus.Logger
	if !ok {
		logger = logrus.New()

		if logger == nil {
			panic("logrus.New() returns nil")
		}

		handler := configureLoggerHandler
		if handler != nil {
			handler(name, logger)
		}

		l, _ = loggers.LoadOrStore(name, logger)
	}
	logger = l.(*logrus.Logger)

	return logger
}

func LoggerEntry(name string) *logrus.Entry {
	logger := getLogger(name)
	fields := logrus.Fields{}
	if name != "" {
		fields["logger"] = name
	}
	if gearModule != "" {
		fields["module"] = gearModule
	}
	return logger.WithFields(fields)
}
