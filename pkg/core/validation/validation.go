package validation

import (
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/louischan-oursky/gojsonschema"
)

func init() {
	gojsonschema.FormatCheckers.Add("URLPathOnly", URL{
		URLVariant: URLVariantPathOnly,
	})
	gojsonschema.FormatCheckers.Add("URLFullOrPath", URL{
		URLVariant: URLVariantFullOrPath,
	})
	gojsonschema.FormatCheckers.Add("URLFullOnly", URL{
		URLVariant: URLVariantFullOrPath,
	})
	gojsonschema.FormatCheckers.Add("RelativeDirectoryPath", FilePath{
		Relative: true,
		File:     false,
	})
	gojsonschema.FormatCheckers.Add("RelativeFilePath", FilePath{
		Relative: true,
		File:     true,
	})
}

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
	return skyerr.NewErrorWithInfo(skyerr.InvalidArgument, message, info)
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

type URLVariant int

const (
	URLVariantFullOnly URLVariant = iota
	URLVariantPathOnly
	URLVariantFullOrPath
)

type URL struct {
	URLVariant URLVariant
}

// nolint: gocyclo
func (f URL) IsFormat(input interface{}) bool {
	str, ok := input.(string)
	if !ok {
		return false
	}
	if str == "" {
		return false
	}

	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return false
	}

	p := ""
	switch f.URLVariant {
	case URLVariantFullOnly:
		if u.Scheme == "" || u.Host == "" {
			return false
		}
		p = u.EscapedPath()
		if p == "" {
			p = "/"
		}
	case URLVariantPathOnly:
		if u.Scheme != "" || u.User != nil || u.Host != "" {
			return false
		}
		p = str
	case URLVariantFullOrPath:
		if u.Scheme != "" || u.User != nil || u.Host != "" {
			p = u.EscapedPath()
			if p == "" {
				p = "/"
			}
		} else {
			p = str
		}
	}

	cleaned := filepath.Clean(p)
	if !filepath.IsAbs(p) || cleaned != p {
		return false
	}

	return true
}

type FilePath struct {
	Relative bool
	File     bool
}

func (f FilePath) IsFormat(input interface{}) bool {
	str, ok := input.(string)
	if !ok {
		return false
	}

	if str == "" {
		return false
	}

	abs := filepath.IsAbs(str)
	if f.Relative && abs {
		return false
	}
	if !f.Relative && !abs {
		return false
	}

	trailingSlash := strings.HasSuffix(str, "/")
	if f.File && trailingSlash {
		return false
	}

	return true
}
