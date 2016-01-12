package plugin

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/robfig/cron"
)

func TestPlugin(t *testing.T) {
	Convey("new plugin from non-registered transport", t, func() {
		defer unregisterAllTransports()

		createPlugin := func() {
			NewPlugin("nonexistent", "/tmp/nonexistent", []string{})
		}
		So(createPlugin, ShouldPanic)
	})

	Convey("new plugin from null transport", t, func() {
		defer unregisterAllTransports()

		RegisterTransport("null", nullFactory{})

		plugin := NewPlugin("null", "/tmp/nonexistent", []string{})
		So(plugin, ShouldHaveSameTypeAs, Plugin{})
		So(plugin.transport, ShouldHaveSameTypeAs, &nullTransport{})
	})

	Convey("panic unable to register timer", t, func() {
		RegisterTransport("null", nullFactory{})
		plugin := NewPlugin("null", "/tmp/nonexistent", []string{})

		c := cron.New()
		panicFunc := func() {
			plugin.initTimer(c, []timerInfo{
				{"timerName", "incorrect-spec"},
			})
		}
		So(panicFunc, ShouldPanic)
	})

}

func TestLambda(t *testing.T) {
	Convey("create simple lambda", t, func() {
		handler := createLambdaHandler(map[string]interface{}{
			"name": "hello:world",
		})

		So(handler.Name, ShouldEqual, "hello:world")
		So(handler.AccessKeyRequired, ShouldBeFalse)
		So(handler.UserRequired, ShouldBeFalse)
	})

	Convey("create user required lambda", t, func() {
		handler := createLambdaHandler(map[string]interface{}{
			"name":          "hello:world",
			"user_required": true,
		})

		So(handler.UserRequired, ShouldBeTrue)
	})

	Convey("create key required lambda", t, func() {
		handler := createLambdaHandler(map[string]interface{}{
			"name":         "hello:world",
			"key_required": true,
		})

		So(handler.AccessKeyRequired, ShouldBeTrue)
	})
}
