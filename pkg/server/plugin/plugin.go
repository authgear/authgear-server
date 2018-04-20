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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/robfig/cron"
	"github.com/skygeario/skygear-server/pkg/server/logging"
	"github.com/skygeario/skygear-server/pkg/server/plugin/hook"
	"github.com/skygeario/skygear-server/pkg/server/plugin/provider"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skyconfig"
)

var log = logging.LoggerEntry("plugin")

const (
	// PluginInitMaxRetryCount defines the maximum retries for plugin initialization
	PluginInitMaxRetryCount = 100
)

// Plugin represents a collection of handlers, hooks and lambda functions
// that extends or modifies functionality provided by skygear.
type Plugin struct {
	initRetryCount int
	transport      Transport
	gatewayMap     map[string]*router.Gateway
}

type pluginHandlerInfo struct {
	AuthRequired bool     `json:"auth_required"`
	Name         string   `json:"name"`
	Methods      []string `json:"methods"`
	KeyRequired  bool     `json:"key_required"`
	UserRequired bool     `json:"user_required"`
}

type pluginHookInfo struct {
	Async   bool   `json:"async"`   // execute hook asynchronously
	Trigger string `json:"trigger"` // before_save etc.
	Type    string `json:"type"`    // record type
	Name    string `json:"name"`    // hook name
}

type timerInfo struct {
	Name string `json:"name"`
	Spec string `json:"spec"`
}

type providerInfo struct {
	Type string `json:"type"`
	Name string `json:"id"`
}

type registrationInfo struct {
	Handlers  []pluginHandlerInfo      `json:"handler"`
	Hooks     []pluginHookInfo         `json:"hook"`
	Lambdas   []map[string]interface{} `json:"op"`
	Timers    []timerInfo              `json:"timer"`
	Providers []providerInfo           `json:"provider"`
}

var transportFactories = map[string]TransportFactory{}

// RegisterTransport registers a transport factory by name.
func RegisterTransport(name string, transport TransportFactory) {
	transportFactories[name] = transport
}

// SupportedTransports tells all supported transport names
func SupportedTransports() []string {
	var transports []string
	for name := range transportFactories {
		transports = append(transports, name)
	}
	return transports
}

func unregisterAllTransports() {
	transportFactories = map[string]TransportFactory{}
}

// NewPlugin creates an instance of Plugin by transport and configuration.
func NewPlugin(name string, path string, args []string, config skyconfig.Configuration) Plugin {
	factory := transportFactories[name]
	if factory == nil {
		panic(fmt.Errorf("unable to find plugin transport '%v'", name))
	}
	p := Plugin{
		transport:  factory.Open(path, args, config),
		gatewayMap: map[string]*router.Gateway{},
	}
	return p
}

// Context contains reference to structs that will be initialized by plugin.
type Context struct {
	plugins          []*Plugin
	Router           *router.Router
	Mux              *http.ServeMux
	HandlerInjector  router.HandlerInjector
	HookRegistry     *hook.Registry
	ProviderRegistry *provider.Registry
	Scheduler        *cron.Cron
	Config           skyconfig.Configuration
}

// AddPluginConfiguration creates and appends a plugin
func (c *Context) AddPluginConfiguration(name string, path string, args []string) *Plugin {
	plug := NewPlugin(name, path, args, c.Config)
	c.plugins = append(c.plugins, &plug)
	return &plug
}

func (c *Context) getInitPayload() ([]byte, error) {
	payload := struct {
		Config skyconfig.Configuration `json:"config"`
	}{c.Config}

	return json.Marshal(payload)
}

// InitPlugins initializes all plugins registered
func (c *Context) InitPlugins() {
	wg := sync.WaitGroup{}
	for _, eachPlugin := range c.plugins {
		wg.Add(1)
		go func(plug *Plugin) {
			defer wg.Done()
			plug.Init(c)
		}(eachPlugin)
	}

	// make a goroutine to wait for all the plugins to be ready
	go func() {
		log.
			WithField("count", len(c.plugins)).
			Info("Wait for all plugin configurations")
		wg.Wait()

		data, err := c.getInitPayload()
		if err != nil {
			log.WithField("error", err).Warning("Fail to get init payload")
			data = []byte{}
		}
		c.SendEvent("before-plugins-ready", data, false)

		for _, eachPlugin := range c.plugins {
			eachPlugin.transport.SetState(TransportStateReady)
		}

		c.SendEvent("after-plugins-ready", data, false)
		c.SendEvent("server-ready", data, false)
	}()
}

// IsInitialized returns true if all the plugins have been initialized
func (c *Context) IsInitialized() bool {
	for _, eachPlugin := range c.plugins {
		if !eachPlugin.IsInitialized() {
			return false
		}
	}
	return true
}

// IsReady returns true if all the plugins are ready for client requests
func (c *Context) IsReady() bool {
	for _, eachPlugin := range c.plugins {
		if !eachPlugin.IsReady() {
			return false
		}
	}
	return true
}

// SendEvent sends event to all plugins
//
// SendEvent accepts `async` flag. Setting `async` to `false` means that
// an event will be sent to a plugin after another.
func (c *Context) SendEvent(name string, data []byte, async bool) {
	sendEventFunc := func(plugin *Plugin, name string, data []byte) {
		plugin.transport.SendEvent(name, data)
	}

	for _, eachPlugin := range c.plugins {
		if async {
			go sendEventFunc(eachPlugin, name, data)
		} else {
			sendEventFunc(eachPlugin, name, data)
		}
	}
}

// Init instantiates a plugin. This sets up hooks and handlers.
func (p *Plugin) Init(context *Context) {
	data, err := context.getInitPayload()
	if err != nil {
		log.WithField("error", err).Panic("Fail to get init payload")
	}

	transport := p.transport
	if bidirectional, ok := transport.(BidirectionalTransport); ok {
		bidirectional.SetRouter(context.Router)
	}

	transport.SendEvent("before-config", data)
	for {
		log.
			WithField("retry", p.initRetryCount).
			Info("Sending init event to plugin")

		transport.SetState(TransportStateUninitialized)
		regInfo, err := p.requestInit(data)
		if err != nil {
			transport.SetState(TransportStateError)

			p.initRetryCount++
			if p.initRetryCount >= PluginInitMaxRetryCount {
				log.Panic("Fail to initialize plugin")
			}
			time.Sleep(2 * time.Second)
			continue
		}

		p.processRegistrationInfo(context, regInfo)
		transport.SetState(TransportStateInitialized)

		break
	}
	transport.SendEvent("after-config", data)
}

func (p *Plugin) requestInit(data []byte) (regInfo registrationInfo, initErr error) {
	out, err := p.transport.SendEvent("init", data)
	log.WithFields(logrus.Fields{
		"out":    string(out),
		"err":    err,
		"plugin": p,
	}).Info("Get response from init")

	if err != nil {
		initErr = fmt.Errorf("Cannot encode plugin initialization payload. Error: %v", err)
		return
	}

	if err := json.Unmarshal(out, &regInfo); err != nil {
		initErr = fmt.Errorf("Unable to decode plugin initialization info. Error: %v", err)
		return
	}

	return
}

// IsInitialized returns true if the plugin has been initialized
func (p *Plugin) IsInitialized() bool {
	transportState := p.transport.State()
	return transportState == TransportStateInitialized ||
		transportState == TransportStateReady
}

// IsReady returns true if the plugin is ready for client request
func (p *Plugin) IsReady() bool {
	return p.transport.State() == TransportStateReady
}

func (p *Plugin) processRegistrationInfo(context *Context, regInfo registrationInfo) {
	log.WithFields(logrus.Fields{
		"regInfo":   regInfo,
		"transport": p.transport,
	}).Debugln("Got configuration from plugin, registering")
	p.initHandler(context.Mux, context.HandlerInjector, regInfo.Handlers, context.Config)
	p.initLambda(context.Router, context.HandlerInjector, regInfo.Lambdas)
	p.initHook(context.HookRegistry, regInfo.Hooks)
	if context.Scheduler != nil {
		p.initTimer(context.Scheduler, regInfo.Timers)
	} else {
		log.Info("Ignoring scheduled cron jobs because server is in slave mode.")
	}
	p.initProvider(context.ProviderRegistry, regInfo.Providers)
}

func (p *Plugin) initHandler(mux *http.ServeMux, injector router.HandlerInjector, handlers []pluginHandlerInfo, config skyconfig.Configuration) {
	for _, handler := range handlers {
		h := NewPluginHandler(handler, p)
		injector.Inject(h)
		h.Setup()
		name := h.Name
		name = strings.Replace(name, ":", "/", -1)
		if !strings.HasPrefix(name, "/") {
			name = "/" + name
		}
		var handlerGateway *router.Gateway
		handlerGateway, ok := p.gatewayMap[name]
		if !ok {
			handlerGateway = router.NewGateway("", name, mux)
			handlerGateway.ResponseTimeout = time.Duration(config.App.ResponseTimeout) * time.Second
			p.gatewayMap[name] = handlerGateway
		}
		for _, method := range handler.Methods {
			handlerGateway.Handle(method, h)
		}
		log.Debugf(`Registered handler "%s" with serveMux at path "%s"`, h.Name, name)
	}
}

func (p *Plugin) initLambda(r *router.Router, injector router.HandlerInjector, lambdas []map[string]interface{}) {
	for _, lambda := range lambdas {
		handler := NewLambdaHandler(lambda, p)
		injector.Inject(handler)
		handler.Setup()
		r.Map(handler.Name, handler)
		log.Debugf(`Registered lambda "%s" with router.`, handler.Name)
	}
}

func (p *Plugin) initHook(registry *hook.Registry, hookInfos []pluginHookInfo) {
	for _, hookInfo := range hookInfos {
		kind := hook.Kind(hookInfo.Trigger)
		recordType := hookInfo.Type

		registry.Register(kind, recordType, CreateHookFunc(p, hookInfo))
	}
}

func (p *Plugin) initTimer(c *cron.Cron, timerInfos []timerInfo) {
	for _, timerInfo := range timerInfos {
		timerName := timerInfo.Name
		err := c.AddFunc(timerInfo.Spec, func() {
			output, _ := p.transport.RunTimer(timerName, []byte{})
			log.Debugf("Executed a timer{%v} with result: %s", timerName, output)
		})

		if err != nil {
			panic(fmt.Errorf(`unable to add timer for "%s": %s`, timerName, err))
		}
	}
}

func (p *Plugin) initProvider(registry *provider.Registry, providerInfos []providerInfo) {
	for _, providerInfo := range providerInfos {
		provider := NewAuthProvider(providerInfo.Name, p)
		registry.RegisterAuthProvider(providerInfo.Name, provider)
	}
}
