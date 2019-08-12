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
	gojsonschema.FormatCheckers.Add("URLPath", URL{
		AllowAbsoluteURI: false,
	})
	gojsonschema.FormatCheckers.Add("RelativeDirectoryPath", FilePath{
		Relative: true,
		File:     false,
	})
	gojsonschema.FormatCheckers.Add("RelativeFilePath", FilePath{
		Relative: true,
		File:     true,
	})
	gojsonschema.FormatCheckers.Add("HookURL", URL{
		AllowAbsoluteURI: true,
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
	for _, cause := range e.Causes {
		arguments = append(arguments, cause.String())
	}
	return skyerr.NewInvalidArgument(message, arguments)
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

type URL struct {
	AllowAbsoluteURI bool
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
	if !f.AllowAbsoluteURI && (u.Scheme != "" || u.User != nil || u.Host != "") {
		return false
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return false
	}

	var p string
	if f.AllowAbsoluteURI {
		p = u.EscapedPath()
	} else {
		p = str
	}
	if p == "" {
		p = "/"
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
