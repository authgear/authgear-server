package plugin

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
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
		So(plugin.transport, ShouldHaveSameTypeAs, nullTransport{})
	})

}
