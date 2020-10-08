package configsource_test

import (
	"io/ioutil"
	"testing"

	corev1 "k8s.io/api/core/v1"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
)

func TestMakeAppFS(t *testing.T) {
	Convey("MakeAppFS", t, func() {
		secret := &corev1.Secret{
			Data: map[string][]byte{
				configsource.EscapePath("authgear.yaml"):              []byte("authgear.yaml"),
				configsource.EscapePath("templates/translation.json"): []byte("templates/translation.json"),
				configsource.EscapePath("authgear.secrets.yaml"):      []byte("authgear.secrets.yaml"),
			},
		}

		fs, err := configsource.MakeAppFS(secret)
		So(err, ShouldBeNil)

		f1, err := fs.Open("templates/translation.json")
		So(err, ShouldBeNil)
		defer f1.Close()

		content, err := ioutil.ReadAll(f1)
		So(err, ShouldBeNil)
		So(content, ShouldResemble, []byte("templates/translation.json"))

		f2, err := fs.Open("/templates/translation.json")
		So(err, ShouldBeNil)
		defer f2.Close()

		content, err = ioutil.ReadAll(f2)
		So(err, ShouldBeNil)
		So(content, ShouldResemble, []byte("templates/translation.json"))
	})
}
