package webapp

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

// RenderProvider renders HTML template.
type RenderProvider interface {
	WritePage(w http.ResponseWriter, r *http.Request, templateType config.TemplateItemType, err error)
}
