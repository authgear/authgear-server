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

package plugin

import (
	"net/http"
	"testing"

	"github.com/facebookgo/inject"
	"github.com/robfig/cron"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skyconfig"
)

type ContextKey string

var HelloContextKey ContextKey = "hello"

type MockPluginReadyPreprocessor struct{}

func (p MockPluginReadyPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	return http.StatusOK
}

type MockInjectAuthIDPreprocessor struct{}

func (p MockInjectAuthIDPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	return http.StatusOK
}

type MockAccessKeyPreprocessor struct{}

func (p MockAccessKeyPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	payload.AccessKey = router.ClientAccessKey
	return http.StatusOK
}

type MockNullPreprocessor struct {
}

func (p MockNullPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	return http.StatusOK
}

func TestPlugin(t *testing.T) {
	config := skyconfig.Configuration{}
	Convey("new plugin from non-registered transport", t, func() {
		defer unregisterAllTransports()

		createPlugin := func() {
			NewPlugin("nonexistent", "/tmp/nonexistent", []string{}, config)
		}
		So(createPlugin, ShouldPanic)
	})

	Convey("new plugin from null transport", t, func() {
		defer unregisterAllTransports()

		RegisterTransport("null", nullFactory{})

		plugin := NewPlugin("null", "/tmp/nonexistent", []string{}, config)
		So(plugin, ShouldHaveSameTypeAs, Plugin{})
		So(plugin.transport, ShouldHaveSameTypeAs, &nullTransport{})
	})

	Convey("panic unable to register timer", t, func() {
		RegisterTransport("null", nullFactory{})
		plugin := NewPlugin("null", "/tmp/nonexistent", []string{}, config)

		c := cron.New()
		panicFunc := func() {
			plugin.initTimer(c, []timerInfo{
				{"timerName", "incorrect-spec"},
			})
		}
		So(panicFunc, ShouldPanic)
	})

	Convey("init handler", t, func() {
		RegisterTransport("null", nullFactory{})
		plugin := NewPlugin("null", "/tmp/nonexistent", []string{}, config)
		Convey("init correctly with one handler", func() {
			mux := http.NewServeMux()
			plugin.initHandler(mux, router.HandlerInjector{
				ServiceGraph: &inject.Graph{},
				PreprocessorMap: &router.PreprocessorRegistry{
					"inject_auth_id": MockInjectAuthIDPreprocessor{},
					"plugin_ready":   MockPluginReadyPreprocessor{},
					"authenticator":  MockNullPreprocessor{},
					"dbconn":         MockNullPreprocessor{},
					"require_auth":   MockNullPreprocessor{},
					"check_user":     MockNullPreprocessor{},
				},
			}, []pluginHandlerInfo{
				pluginHandlerInfo{
					Name: "chima:echo",
				},
			}, config)
			So(len(plugin.gatewayMap), ShouldEqual, 1)
			So(plugin.gatewayMap, ShouldContainKey, "/chima/echo")
		})

		Convey("init correctly with multiple handler", func() {
			mux := http.NewServeMux()
			plugin.initHandler(mux, router.HandlerInjector{
				ServiceGraph: &inject.Graph{},
				PreprocessorMap: &router.PreprocessorRegistry{
					"inject_auth_id": MockInjectAuthIDPreprocessor{},
					"plugin_ready":   MockPluginReadyPreprocessor{},
					"authenticator":  MockNullPreprocessor{},
					"dbconn":         MockNullPreprocessor{},
					"require_auth":   MockNullPreprocessor{},
					"check_user":     MockNullPreprocessor{},
				},
			}, []pluginHandlerInfo{
				pluginHandlerInfo{
					Name: "chima:echo",
				},
				pluginHandlerInfo{
					Name:    "faseng:location",
					Methods: []string{"GET"},
				},
				pluginHandlerInfo{
					Name:    "faseng:location",
					Methods: []string{"POST", "PUT"},
				},
			}, config)
			So(len(plugin.gatewayMap), ShouldEqual, 2)
			So(plugin.gatewayMap, ShouldContainKey, "/chima/echo")
			So(plugin.gatewayMap, ShouldContainKey, "/faseng/location")
		})
	})

}
