package plugin

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/oursky/skygear/hook"
	"github.com/oursky/skygear/provider"
	"github.com/oursky/skygear/router"
	"github.com/robfig/cron"
)

// Plugin represents a collection of handlers, hooks and lambda functions
// that extends or modifies functionality provided by skygear.
type Plugin struct {
	transport Transport
}

type pluginHandlerInfo struct {
	AuthRequired bool `json:"auth_required"`
}

type pluginHookInfo struct {
	Async   bool   `json:"async"`
	Trigger string `json:"trigger"`
	Type    string `json:"type"`
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
	Handlers  map[string]pluginHandlerInfo `json:"handler"`
	Hooks     []pluginHookInfo             `json:"hook"`
	Lambdas   []string                     `json:"op"`
	Timers    []timerInfo                  `json:"timer"`
	Providers []providerInfo               `json:"provider"`
}

func (p *Plugin) getRegistrationInfo() registrationInfo {
	outBytes, err := p.transport.RunInit()
	if err != nil {
		panic(fmt.Sprintf("Unable to get registration info from plugin. Error: %v", err))
	}

	regInfo := registrationInfo{}
	if err := json.Unmarshal(outBytes, &regInfo); err != nil {
		panic(err)
	}
	return regInfo
}

var transportFactories = map[string]TransportFactory{}

// RegisterTransport registers a transport factory by name.
func RegisterTransport(name string, transport TransportFactory) {
	transportFactories[name] = transport
}

func unregisterAllTransports() {
	transportFactories = map[string]TransportFactory{}
}

// NewPlugin creates an instance of Plugin by transport and configuration.
func NewPlugin(name string, path string, args []string) Plugin {
	factory := transportFactories[name]
	if factory == nil {
		panic(fmt.Errorf("unable to find plugin transport '%v'", name))
	}
	p := Plugin{
		transport: factory.Open(path, args),
	}
	return p
}

// InitContext contains reference to structs that will be initialized by plugin.
type InitContext struct {
	Router           *router.Router
	HookRegistry     *hook.Registry
	ProviderRegistry *provider.Registry
	Scheduler        *cron.Cron
}

// Init instantiates a plugin. This sets up hooks and handlers.
func (p *Plugin) Init(context *InitContext) {
	regInfo := p.getRegistrationInfo()

	log.WithFields(log.Fields{
		"regInfo":   regInfo,
		"transport": p.transport,
	}).Debugln("Got configuration from pligin, registering")
	p.initLambda(context.Router, regInfo.Lambdas)
	p.initHook(context.HookRegistry, regInfo.Hooks)
	p.initTimer(context.Scheduler, regInfo.Timers)
	p.initProvider(context.ProviderRegistry, regInfo.Providers)
}

func (p *Plugin) initLambda(r *router.Router, lambdaNames []string) {
	for _, lambdaName := range lambdaNames {
		r.Map(lambdaName, CreateLambdaHandler(p, lambdaName))
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
