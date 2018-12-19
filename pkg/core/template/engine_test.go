package template

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type MockDefaultLoader struct{}

func (m *MockDefaultLoader) Load(name string) (string, error) {
	if name == "name1" || name == "name2" {
		return fmt.Sprintf("default content from %s", name), nil
	}

	return "", errors.New("template not found")
}

type MockLoader struct{}

func (m *MockLoader) Load(name string) (string, error) {
	if name == "name1" || name == "name3" {
		return fmt.Sprintf("load content from %s", name), nil
	}

	return "", errors.New("template not found")
}
func TestEngine(t *testing.T) {
	engine := Engine{
		defaultLoader: &StringLoader{
			StringMap: map[string]string{
				"name1": "default content from name1",
				"name2": "default content from name2",
			},
		},
		loaders: []Loader{
			&MockLoader{},
		},
	}

	emptyContext := map[string]interface{}{}

	Convey("return content from loader when both exist", t, func() {
		out, err := engine.ParseTextTemplate("name1", emptyContext, ParseOption{Required: true})
		So(err, ShouldBeNil)
		So(out, ShouldEqual, "load content from name1")
	})

	Convey("return content from default if loader does not have template", t, func() {
		out, err := engine.ParseTextTemplate("name2", emptyContext, ParseOption{Required: true})
		So(err, ShouldBeNil)
		So(out, ShouldEqual, "default content from name2")

		out, err = engine.ParseTextTemplate("name2", emptyContext, ParseOption{Required: true, FallbackTemplateName: "name1"})
		So(err, ShouldBeNil)
		So(out, ShouldEqual, "default content from name2")
	})

	Convey("return content from fallback if loader does not have template", t, func() {
		out, err := engine.ParseTextTemplate("name4", emptyContext, ParseOption{Required: true, FallbackTemplateName: "name1"})
		So(err, ShouldBeNil)
		So(out, ShouldEqual, "load content from name1")

		out, err = engine.ParseTextTemplate("name4", emptyContext, ParseOption{Required: true, FallbackTemplateName: "name2"})
		So(err, ShouldBeNil)
		So(out, ShouldEqual, "default content from name2")
	})

	Convey("return content from loader even default not set", t, func() {
		out, err := engine.ParseTextTemplate("name3", emptyContext, ParseOption{Required: true})
		So(err, ShouldBeNil)
		So(out, ShouldEqual, "load content from name3")
	})

	Convey("return empty string for name not registered and required is false", t, func() {
		out, err := engine.ParseTextTemplate("random", emptyContext, ParseOption{Required: false})
		So(err, ShouldBeNil)
		So(out, ShouldEqual, "")
	})

	Convey("panic for name not registered and required is true", t, func() {
		_, err := engine.ParseTextTemplate("random", emptyContext, ParseOption{Required: true})
		So(err, ShouldNotBeNil)

		_, err = engine.ParseTextTemplate("random", emptyContext, ParseOption{Required: true, FallbackTemplateName: "random2"})
		So(err, ShouldNotBeNil)
	})
}
