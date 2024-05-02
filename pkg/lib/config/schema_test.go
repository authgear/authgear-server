package config_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	goyaml "gopkg.in/yaml.v2"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestDumpSchema(t *testing.T) {
	Convey("DumpSchema", t, func() {
		s, err := config.DumpSchema()
		So(err, ShouldBeNil)
		t.Logf("Dumping the schema of authgear.yaml\n%s", s)
	})
}

func TestDumpSecretConfigSchema(t *testing.T) {
	Convey("DumpSecretConfigSchema", t, func() {
		s, err := config.DumpSecretConfigSchema()
		So(err, ShouldBeNil)
		t.Logf("Dumping the schema of authgear.secret.yaml\n%s", s)
	})
}

func TestAppConfigSchema(t *testing.T) {
	testFiles := []string{
		"testdata/hook_tests.yaml",
		"testdata/custom_attributes_tests.yaml",
		"testdata/account_deletion_tests.yaml",
		"testdata/account_anonymization_tests.yaml",
		"testdata/google_tag_manager_tests.yaml",
		"testdata/rate_limit_tests.yaml",
		"testdata/authentication_flow_object_name_tests.yaml",
		"testdata/identification_method_tests.yaml",
		"testdata/signup_flow_tests.yaml",
		"testdata/login_flow_tests.yaml",
		"testdata/signup_login_flow_tests.yaml",
		"testdata/reauth_flow_tests.yaml",
		"testdata/account_linking_tests.yaml",
	}

	type TestCase struct {
		Part  string      `yaml:"part"`
		Name  string      `yaml:"name"`
		Error *string     `yaml:"error"`
		Value interface{} `yaml:"value"`
	}
	var testCases []TestCase
	loadTestCases := func(filename string) {
		f, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		decoder := goyaml.NewDecoder(f)
		for {
			var testCase TestCase
			err := decoder.Decode(&testCase)
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				panic(err)
			}

			testCases = append(testCases, testCase)
		}
	}
	for _, n := range testFiles {
		loadTestCases(n)
	}

	Convey("AppConfig schema parts", t, func() {
		for _, testCase := range testCases {
			name := fmt.Sprintf("%s/%s", testCase.Part, testCase.Name)
			Convey(name, func() {
				data, err := goyaml.Marshal(testCase.Value)
				if err != nil {
					panic(err)
				}
				data, err = yaml.YAMLToJSON(data)
				if err != nil {
					panic(err)
				}

				err = config.Schema.PartValidator(testCase.Part).Validate(bytes.NewReader(data))
				if testCase.Error != nil {
					So(err, ShouldBeError, *testCase.Error)
				} else {
					So(err, ShouldBeNil)
				}
			})
		}
	})
}
