package plugin

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/oursky/ourd/hook"
	"github.com/oursky/ourd/router"
	"github.com/robfig/cron"
)

// Plugin represents a collection of handlers, hooks and lambda functions
// that extends or modifies functionality provided by ourd.
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

type registrationInfo struct {
	Handlers map[string]pluginHandlerInfo `json:"handler"`
	Hooks    []pluginHookInfo             `json:"hook"`
	Lambdas  []string                     `json:"op"`
	Timers   []timerInfo                  `json:"timer"`
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

// Init instantiates a plugin. This sets up hooks and handlers.
func (p *Plugin) Init(r *router.Router, registry *hook.Registry, c *cron.Cron) {
	regInfo := p.getRegistrationInfo()

	log.WithFields(log.Fields{
		"regInfo":   regInfo,
		"transport": p.transport,
	}).Debugln("Got configuration from pligin, registering")
	p.initLambda(r, regInfo.Lambdas)
	p.initHook(registry, regInfo.Hooks)
	p.initTimer(c, regInfo.Timers)
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
		c.AddFunc(timerInfo.Spec, func() {
			output, _ := p.transport.RunTimer(timerName, []byte{})
			log.Debugf("Executed a timer{%v} with result: %s", timerName, output)
		})
	}
}
