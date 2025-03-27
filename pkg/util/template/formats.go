package template

import (
	"context"
	"fmt"
	texttemplate "text/template"

	jsonschemaformat "github.com/iawaknahc/jsonschema/pkg/jsonschema/format"
)

func init() {
	jsonschemaformat.DefaultChecker["x_text_template"] = FormatTextTemplate{}
}

type FormatTextTemplate struct{}

func (FormatTextTemplate) CheckFormat(ctx context.Context, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	t, err := texttemplate.New("").Parse(str)
	if err != nil {
		return fmt.Errorf("invalid text template")
	}

	agtpl := &AGTextTemplate{}
	return agtpl.Wrap(t)
}
