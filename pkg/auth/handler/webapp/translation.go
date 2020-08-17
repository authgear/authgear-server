package webapp

import (
	"github.com/authgear/authgear-server/pkg/util/template"
)

const (
	TemplateItemTypeAuthUITranslationJSON string = "auth_ui_translation.json"
)

var TemplateAuthUITranslationJSON = template.T{
	Type: TemplateItemTypeAuthUITranslationJSON,
}
