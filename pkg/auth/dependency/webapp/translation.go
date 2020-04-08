package webapp

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

const (
	TemplateItemTypeAuthUITranslationJSON config.TemplateItemType = "auth_ui_translation.json"
)

var TemplateAuthUITranslationJSON = template.Spec{
	Type: TemplateItemTypeAuthUITranslationJSON,
	Default: `
{
}
	`,
}
