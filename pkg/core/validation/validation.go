package validation

import (
	"fmt"
	"sort"
	"strings"

	"github.com/xeipuuv/gojsonschema"

	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type Cause struct {
	Message     string
	InstancePtr string
}

func (c Cause) String() string {
	return fmt.Sprintf("%s: %s", c.InstancePtr, c.Message)
}

type Error struct {
	Causes []Cause
}

func (e Error) Error() string {
	w := &strings.Builder{}
	for _, cause := range e.Causes {
		fmt.Fprintf(w, "%s\n", cause.String())
	}
	return w.String()
}

func ConvertErrors(errs []gojsonschema.ResultError) error {
	var causes []Cause
	for _, e := range errs {
		causes = append(causes, Cause{
			InstancePtr: fmt.Sprintf("#%s", e.Context().JSONPointer()),
			Message:     e.Description(),
		})
	}
	err := Error{
		Causes: causes,
	}
	sort.Sort(err)
	return err
}

func (e Error) SkyErrInvalidArgument(message string) error {
	var arguments []string
	var arr []map[string]interface{}
	for _, cause := range e.Causes {
		arguments = append(arguments, cause.String())
		arr = append(arr, map[string]interface{}{
			"pointer": cause.InstancePtr,
			"message": cause.Message,
		})
	}
	info := map[string]interface{}{
		"arguments": arguments,
		"causes":    arr,
	}
	// TODO(error): JSON schema
	return errors.WithDetails(skyerr.NewInvalid(message), info)
}

func (e Error) Len() int {
	return len(e.Causes)
}

func (e Error) Less(i, j int) bool {
	ci := e.Causes[i]
	cj := e.Causes[j]
	if ci.InstancePtr < cj.InstancePtr {
		return true
	} else if ci.InstancePtr == cj.InstancePtr {
		return ci.Message < cj.Message
	}
	return false
}

func (e Error) Swap(i, j int) {
	e.Causes[i], e.Causes[j] = e.Causes[j], e.Causes[i]
}
