package template

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func MockDownloadFromFilePath(filepath string) (string, error) {
	if filepath == "filepath1" || filepath == "filepath2" || filepath == "filepath3" {
		return fmt.Sprintf("content from %s", filepath), nil
	}

	return "", errors.New("file not found")
}

func MockDownloadFromURL(url string) (string, error) {
	if url == "url1" {
		return fmt.Sprintf("content from %s", url), nil
	}

	return "", errors.New("file not found")
}

func TestEngine(t *testing.T) {
	OriginalDownloadFromFilePath := downloadFromFilePath
	downloadFromFilePath = MockDownloadFromFilePath
	defer func() {
		downloadFromFilePath = OriginalDownloadFromFilePath
	}()

	OriginalDownloadFromURL := downloadFromURL
	downloadFromURL = MockDownloadFromURL
	defer func() {
		downloadFromURL = OriginalDownloadFromURL
	}()

	engine := Engine{
		defaultPathMap: map[string]string{
			"name1": "filepath1",
			"name2": "filepath2",
			"name3": "filepath3",
		},
		urlMap: map[string]string{
			"name1":      "url1",
			"name2":      "url2",
			"no-content": "no-content",
		},
	}

	emptyContext := map[string]interface{}{}

	Convey("return content from url when both exist", t, func() {
		out, err := engine.ParseTextTemplate("name1", emptyContext, true)
		So(err, ShouldBeNil)
		So(out, ShouldEqual, "content from url1")
	})

	Convey("return content from file if url exist but throw error", t, func() {
		out, err := engine.ParseTextTemplate("name2", emptyContext, true)
		So(err, ShouldBeNil)
		So(out, ShouldEqual, "content from filepath2")
	})

	Convey("return content from file if url does not exist", t, func() {
		out, err := engine.ParseTextTemplate("name3", emptyContext, true)
		So(err, ShouldBeNil)
		So(out, ShouldEqual, "content from filepath3")
	})

	Convey("return empty string for name not registered and required is false", t, func() {
		out, err := engine.ParseTextTemplate("random", emptyContext, false)
		So(err, ShouldBeNil)
		So(out, ShouldEqual, "")
	})

	Convey("return empty string for name registered but content not found and required is false", t, func() {
		out, err := engine.ParseTextTemplate("no-content", emptyContext, false)
		So(err, ShouldBeNil)
		So(out, ShouldEqual, "")
	})

	Convey("panic for name registered but content not found and required is true", t, func() {
		So(func() {
			engine.ParseTextTemplate("no-content", emptyContext, true)
		}, ShouldPanic)
	})

	Convey("panic for name not registered and required is true", t, func() {
		So(func() {
			engine.ParseTextTemplate("random", emptyContext, true)
		}, ShouldPanic)
	})
}
