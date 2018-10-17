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
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	loggers                map[string]*logrus.Logger
	lock                   sync.Mutex
	configureLoggerHandler func(string, *logrus.Logger)
	gearModule             string
)

func init() {
	loggers = map[string]*logrus.Logger{}
	loggers[""] = logrus.StandardLogger()
}

func SetConfigureLoggerHandler(handler func(string, *logrus.Logger)) {
	configureLoggerHandler = handler
}

func SetModule(module string) {
	gearModule = module
}

func Logger(name string) *logrus.Logger {
	lock.Lock()
	defer lock.Unlock()

	logger, ok := loggers[name]
	if !ok {
		logger = logrus.New()

		if logger == nil {
			panic("logrus.New() returns nil")
		}

		handler := configureLoggerHandler
		if handler != nil {
			handler(name, logger)
		}

		loggers[name] = logger
	}

	return logger
}

func Loggers() map[string]*logrus.Logger {
	lock.Lock()
	defer lock.Unlock()

	ret := map[string]*logrus.Logger{}
	for loggerName, logger := range loggers {
		ret[loggerName] = logger
	}
	return ret
}

func LoggerEntry(name string) *logrus.Entry {
	return LoggerEntryWithTag(name, name)
}

func LoggerEntryWithTag(name string, tag string) *logrus.Entry {
	logger := Logger(name)
	fields := logrus.Fields{}
	if name != "" {
		fields["logger"] = name
	}
	if tag != "" {
		fields["tag"] = tag
	}
	if gearModule != "" {
		fields["module"] = gearModule
	}
	return logger.WithFields(fields)
}

func CreateLogger(r *http.Request, logger string) *logrus.Entry {
	fields := logrus.Fields{}
	if requestID := r.Header.Get("X-Skygear-Request-ID"); requestID != "" {
		fields["request_id"] = requestID
	}
	return LoggerEntry(logger).WithFields(fields)
}
