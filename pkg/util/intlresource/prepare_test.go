package intlresource

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestPrepare(t *testing.T) {
	Convey("Prepare", t, func() {
		resources := []resource.ResourceFile{
			resource.ResourceFile{
				Location: resource.Location{
					Path: "/en",
				},
				Data: []byte("en"),
			},
		}
		view := resource.EffectiveResource{
			SupportedTags: []string{"zh"},
			DefaultTag:    "zh",
			PreferredTags: []string{"ja"},
		}

		extractLanguageTag := func(resrc resource.ResourceFile) string {
			return string(resrc.Data)
		}

		bag := make(map[string]resource.ResourceFile)
		add := func(langTag string, resrc resource.ResourceFile) error {
			bag[langTag] = resrc
			return nil
		}

		err := Prepare(resources, view, extractLanguageTag, add)
		So(err, ShouldBeNil)
		So(bag, ShouldResemble, map[string]resource.ResourceFile{
			"en": resource.ResourceFile{
				Location: resource.Location{
					Path: "/en",
				},
				Data: []byte("en"),
			},
		})
	})
}
