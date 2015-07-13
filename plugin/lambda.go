package plugin

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"

	"github.com/oursky/ourd/oderr"
	"github.com/oursky/ourd/router"
)

// Executes lambda function implemented by the plugin.
func lambdaHandler(p *Plugin, l string, payload *router.Payload, response *router.Response) {
	inbytes, err := json.Marshal(payload.Data)
	if err != nil {
		response.Err = oderr.NewUnknownErr(err)
		return
	}

	outbytes, err := p.transport.RunLambda(l, inbytes)
	if err != nil {
		response.Err = oderr.NewUnknownErr(err)
		return
	}

	result := map[string]interface{}{}
	err = json.Unmarshal(outbytes, &result)
	if err != nil {
		response.Err = oderr.NewUnknownErr(err)
		return
	}
	log.WithFields(log.Fields{
		"name":   l,
		"input":  payload.Data,
		"result": result,
		"err":    err,
	}).Debugf("Executed a lambda with result")

	response.Result = result
}

// CreateLambdaHandler creates a router.Handler for the specified lambda function
// implemented by the plugin.
func CreateLambdaHandler(p *Plugin, l string) router.Handler {
	return func(payload *router.Payload, response *router.Response) {
		lambdaHandler(p, l, payload, response)
	}
}
