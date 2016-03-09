package plugin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/oursky/skygear/plugin/hook"
	"github.com/oursky/skygear/plugin/provider"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skyconfig"
	"github.com/robfig/cron"
)

// Plugin represents a collection of handlers, hooks and lambda functions
// that extends or modifies functionality provided by skygear.
type Plugin struct {
	transport Transport
}

type pluginHandlerInfo struct {
	AuthRequired bool     `json:"auth_required"`
	Name         string   `json:"name"`
	Method       []string `json:"method"`
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
		transport: factory.Open(path, args, config),
	}
	return p
}

// InitContext contains reference to structs that will be initialized by plugin.
type InitContext struct {
	plugins          []*Plugin
	Router           *router.Router
	Mux              *http.ServeMux
	Preprocessors    router.PreprocessorRegistry
	HookRegistry     *hook.Registry
	ProviderRegistry *provider.Registry
	Scheduler        *cron.Cron
	Config           skyconfig.Configuration
}

func (c *InitContext) AddPluginConfiguration(name string, path string, args []string) *Plugin {
	plug := NewPlugin(name, path, args, c.Config)
	c.plugins = append(c.plugins, &plug)
	return &plug
}

func (c *InitContext) InitPlugins() {
	for _, plug := range c.plugins {
		plug.Init(c)
	}
}

// IsReady returns true if all the configured plugins are available
func (c *InitContext) IsReady() bool {
	for _, plug := range c.plugins {
		if !plug.IsReady() {
			return false
		}
	}
	return true
}

// Init instantiates a plugin. This sets up hooks and handlers.
func (p *Plugin) Init(context *InitContext) {
	p.transport.SetInitHandler(func(out []byte, err error) error {
		if err != nil {
			panic(fmt.Sprintf("Unable to get registration info from plugin. Error: %v", err))
		}

		regInfo := registrationInfo{}
		if err := json.Unmarshal(out, &regInfo); err != nil {
			panic(err)
		}

		p.processRegistrationInfo(context, regInfo)
		return nil
	})

	log.WithFields(log.Fields{
		"plugin": p,
	}).Debugln("request plugin to return configuration")
	go p.transport.RequestInit()
}

func (p *Plugin) IsReady() bool {
	return p.transport.State() == TransportStateReady
}

func (p *Plugin) processRegistrationInfo(context *InitContext, regInfo registrationInfo) {
	log.WithFields(log.Fields{
		"regInfo":   regInfo,
		"transport": p.transport,
	}).Debugln("Got configuration from pligin, registering")
	p.initHandler(context.Mux, context.Preprocessors, regInfo.Handlers)
	p.initLambda(context.Router, context.Preprocessors, regInfo.Lambdas)
	p.initHook(context.HookRegistry, regInfo.Hooks)
	p.initTimer(context.Scheduler, regInfo.Timers)
	p.initProvider(context.ProviderRegistry, regInfo.Providers)
}

func (p *Plugin) initHandler(mux *http.ServeMux, ppreg router.PreprocessorRegistry, handlers []pluginHandlerInfo) {
	for _, handler := range handlers {
		h := NewPluginHandler(handler, ppreg, p)
		h.Setup()
		name := h.Name
		name = strings.Replace(name, ":", "/", -1)
		if !strings.HasPrefix(name, "/") {
			name = "/" + name
		}
		handlerGateway := router.NewGateway("", name, mux)
		for _, method := range handler.Method {
			if method == "GET" {
				handlerGateway.GET(h)
			}
			if method == "POST" {
				handlerGateway.POST(h)
			}
			if method == "PUT" {
				handlerGateway.PUT(h)
			}
		}
		log.Debugf(`Registered handler "%s" with serveMux at path "%s"`, h.Name, name)
	}
}

func (p *Plugin) initLambda(r *router.Router, ppreg router.PreprocessorRegistry, lambdas []map[string]interface{}) {
	for _, lambda := range lambdas {
		handler := NewLambdaHandler(lambda, ppreg, p)
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
