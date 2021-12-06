package customattrs

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/jsonpointerutil"
)

type Service struct {
	Config *config.CustomAttributesConfig
}

func (s *Service) FromStorageForm(storageForm map[string]interface{}) (T, error) {
	out := make(T)
	for _, c := range s.Config.Attributes {
		ptr, err := jsonpointer.Parse(c.Pointer)
		if err != nil {
			return nil, err
		}

		if val, ok := storageForm[c.ID]; ok {
			err = jsonpointerutil.AssignToJSONObject(ptr, out, val)
			if err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}
