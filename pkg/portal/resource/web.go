package resource

import (
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

var ThemesJSON = PortalRegistry.Register(&resource.SimpleFile{
	Name: "themes.json",
	ParseFn: func(data []byte) (interface{}, error) {
		var parsed interface{}
		err := json.Unmarshal(data, &parsed)
		if err != nil {
			return nil, err
		}
		return parsed, nil
	},
})

var TranslationsJSON = PortalRegistry.Register(&resource.SimpleFile{
	Name: "translations.json",
	ParseFn: func(data []byte) (interface{}, error) {
		var parsed interface{}
		err := json.Unmarshal(data, &parsed)
		if err != nil {
			return nil, err
		}
		return parsed, nil
	},
})
