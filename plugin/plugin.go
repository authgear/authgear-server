package plugin

import (
	"encoding/json"
	"fmt"
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
func (p *Plugin) Init(r *router.Router, c *cron.Cron) {
	regInfo := p.getRegistrationInfo()

	// Initialize lambdas
	for _, lambdaName := range regInfo.Lambdas {
		r.Map(lambdaName, CreateLambdaHandler(p, lambdaName))
	}

	// Initialize timers
	for _, timerInfo := range regInfo.Timers {
		timerName := timerInfo.Name
		c.AddFunc(timerInfo.Spec, func() {
			p.transport.RunTimer(timerName, []byte{})
		})
	}
}
