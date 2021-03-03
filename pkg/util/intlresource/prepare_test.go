package intlresource

import (
	"errors"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

type testFS struct {
	FsLevel resource.FsLevel
}

func (f testFS) Open(name string) (resource.File, error) {
	return nil, errors.New("not implemented")
}

func (f testFS) Stat(name string) (os.FileInfo, error) {
	return nil, errors.New("not implemented")
}

func (f testFS) GetFsLevel() resource.FsLevel {
	return f.FsLevel
}

func TestPrepare(t *testing.T) {
	Convey("Prepare", t, func() {
		resources := []resource.ResourceFile{
			// The builtin resource in intl.DefaultLanguage
			{
				Location: resource.Location{
					Fs:   testFS{resource.FsLevelBuiltin},
					Path: "/en",
				},
				Data: []byte("en"),
			},
			// The resource in fallback language.
			{
				Location: resource.Location{
					Fs:   testFS{resource.FsLevelApp},
					Path: "/zh",
				},
				Data: []byte("zh"),
			},
			// The resource in non-fallback language.
			{
				Location: resource.Location{
					Fs:   testFS{resource.FsLevelApp},
					Path: "/ko",
				},
				Data: []byte("ko"),
			},
		}

		view := resource.EffectiveResource{
			SupportedTags: []string{"zh", "ko"},
			DefaultTag:    "zh",
			PreferredTags: []string{"ja"},
		}

		extractLanguageTag := func(resrc resource.ResourceFile) string {
			return strings.TrimPrefix(resrc.Location.Path, "/")
		}

		bag := make(map[string][]byte)
		add := func(langTag string, resrc resource.ResourceFile) error {
			value := bag[langTag]
			value = append(value, resrc.Data...)
			bag[langTag] = value
			return nil
		}

		err := Prepare(resources, view, extractLanguageTag, add)
		So(err, ShouldBeNil)
		So(bag, ShouldResemble, map[string][]byte{
			"zh": []byte("enzh"),
			"ko": []byte("ko"),
		})
	})
}
