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

package audit

import (
	"fmt"
	"github.com/evalphobia/logrus_fluent"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strconv"
)

const (
	enabledLevel  = logrus.InfoLevel
)

type Event int

const (
	_ Event = iota

	// EventLoginSuccess represents Login Success
	EventLoginSuccess

	// EventLoginFailure represents Login Failure
	EventLoginFailure

	// EventLogout represents Logout
	EventLogout

	// EventSignup represents Signup
	EventSignup

	// EventChangePassword represents Change Password
	EventChangePassword

	// EventChangeRoles represents Change Roles
	EventChangeRoles

	// EventResetPassword represents Reset Password
	EventResetPassword

	// EventDisableUser represents Disable User
	EventDisableUser

	// EventEnableUser represents Enable User
	EventEnableUser
)

func (e Event) String() string {
	switch e {
	case EventLoginSuccess:
		return "login_success"
	case EventLoginFailure:
		return "login_failure"
	case EventLogout:
		return "logout"
	case EventSignup:
		return "signup"
	case EventChangePassword:
		return "change_password"
	case EventChangeRoles:
		return "change_roles"
	case EventResetPassword:
		return "reset_password"
	case EventDisableUser:
		return "disable_user"
	case EventEnableUser:
		return "enable_user"
	default:
		return ""
	}
}

type Trail struct {
	logger *logrus.Entry
}

func (t Trail) Log(entry Entry) {
	t.logger.WithFields(entry.toLogrusFields()).Info("audit_trail")
}

type Entry struct {
	Event         Event
	AuthID        string
	Data          map[string]interface{}
}

func (e *Entry) toLogrusFields() logrus.Fields {
	return logrus.Fields{
		"event":                e.Event.String(),
		"auth_id":              e.AuthID,
		"data":                 e.Data,
	}
}

func createLFSHook(parsedURL *url.URL) (logrus.Hook, error) {
	if parsedURL.Host != "" || parsedURL.Path == "" {
		return nil, fmt.Errorf("malformed file url: %v", parsedURL)
	}
	pathMap := lfshook.PathMap{
		enabledLevel: parsedURL.Path,
	}
	hook := lfshook.NewHook(
		pathMap,
		&logrus.JSONFormatter{},
	)
	return hook, nil
}

func createFluentdHook(parsedURL *url.URL) (logrus.Hook, error) {
	hostname := parsedURL.Hostname()
	if hostname == "" {
		return nil, fmt.Errorf("malformed fluentd url: %v", parsedURL)
	}

	portString := parsedURL.Port()
	var port int
	if portString != "" {
		p, err := strconv.Atoi(portString)
		if err != nil {
			return nil, err
		}
		port = p
	} else {
		port = 24224
	}

	hook, err := logrus_fluent.NewWithConfig(logrus_fluent.Config{
		Host:         hostname,
		Port:         port,
		AsyncConnect: true,
	})
	if err != nil {
		return nil, err
	}
	hook.SetLevels([]logrus.Level{enabledLevel})
	hook.SetTag("skygear.audit")
	return hook, nil
}

func createHook(handlerURL string) (logrus.Hook, error) {
	parsedURL, err := url.Parse(handlerURL)
	if err != nil {
		return nil, err
	}
	scheme := parsedURL.Scheme
	switch scheme {
	case "file":
		return createLFSHook(parsedURL)
	case "fluentd":
		return createFluentdHook(parsedURL)
	}
	return nil, fmt.Errorf("unknown handler: %v, %v", scheme, handlerURL)
}

func NewTrail(handlerURL string, req *http.Request) (*Trail, error) {
	var trailLogger = logrus.New()
	trailLogger.Formatter = &logrus.JSONFormatter{}
	trailLogger.Level = enabledLevel
	if handlerURL != "" {
		hook, err := createHook(handlerURL)
		if err != nil {
			return nil, err
		}
		if hook != nil {
			trailLogger.Hooks.Add(hook)
		}
	}

	fields := logrus.Fields{}
	fields["remote_addr"] = req.RemoteAddr
	fields["x_forwarded_for"] = req.Header.Get("x-forwarded-for")
	fields["x_real_ip"] = req.Header.Get("x-real-ip")
	fields["forwarded"] = req.Header.Get("forwarded")

	return &Trail{
		logger: trailLogger.WithFields(fields),
	}, nil
}
