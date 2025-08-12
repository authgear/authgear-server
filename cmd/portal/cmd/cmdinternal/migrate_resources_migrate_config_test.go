package cmdinternal_test

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	goyaml "go.yaml.in/yaml/v3"

	"github.com/authgear/authgear-server/cmd/portal/cmd/cmdinternal"
)

func TestMigrateConfig(t *testing.T) {
	test := func(name string, fn func(map[string]any) error) {
		f, err := os.Open(path.Join("testdata", name+".yaml"))
		if err != nil {
			panic(err)
		}
		defer f.Close()

		decoder := goyaml.NewDecoder(f)

		var documents []map[string]any
		for {
			var doc map[string]any
			err := decoder.Decode(&doc)
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				panic(err)
			}

			jsonText, err := json.Marshal(doc)
			if err != nil {
				panic(err)
			}
			if err := json.Unmarshal(jsonText, &doc); err != nil {
				panic(err)
			}
			documents = append(documents, doc)
		}
		if len(documents) != 2 {
			panic("unexpected test case format: " + name)
		}

		Convey(name, func() {
			err := fn(documents[0])
			So(err, ShouldBeNil)
			So(documents[0], ShouldResemble, documents[1])
		})
	}

	Convey("migrate configs", t, func() {
		test("rate_limits__migrate_old_config", cmdinternal.MigrateConfigRateLimits)
		test("rate_limits__keep_new_config_if_exist", cmdinternal.MigrateConfigRateLimits)
		test("rate_limits__default_values", cmdinternal.MigrateConfigRateLimits)
	})
}
