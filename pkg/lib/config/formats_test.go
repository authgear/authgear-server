package config

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func makeAppContext(ctx context.Context, cfgYAML []byte) context.Context {
	cfg, err := Parse(ctx, []byte(cfgYAML))
	So(err, ShouldBeNil)

	appCtx := &AppContext{
		Config: &Config{AppConfig: cfg},
	}
	ctx = WithAppContext(ctx, appCtx)
	return ctx
}

func TestFormatPhone(t *testing.T) {

	Convey("FormatPhone", t, func() {
		f := FormatPhone{}.CheckFormat
		So(f(context.Background(), 1), ShouldBeNil)
		So(f(context.Background(), "+85298765432"), ShouldBeNil)
		So(f(context.Background(), ""), ShouldBeError, "not in E.164 format")
		So(f(context.Background(), "foobar"), ShouldBeError, "not in E.164 format")
	})

	Convey("FormatPhone - libphonenumber:isPossibleNumber", t, func() {
		cfgYAML := `
id: test
http:
  public_origin: http://test
ui:
  phone_input:
    validation:
      implementation: libphonenumber
      libphonenumber:
        validation_method: isPossibleNumber
    `
		ctx := context.Background()
		ctx = makeAppContext(ctx, []byte(cfgYAML))

		f := FormatPhone{}.CheckFormat
		So(f(ctx, 1), ShouldBeNil)
		So(f(ctx, "+85212341234"), ShouldBeNil)
		So(f(ctx, ""), ShouldBeError, "not in E.164 format")
		So(f(ctx, "foobar"), ShouldBeError, "not in E.164 format")
	})

	Convey("FormatPhone - libphonenumber:isValidNumber", t, func() {
		cfgYAML := `
id: test
http:
  public_origin: http://test
ui:
  phone_input:
    validation:
      implementation: libphonenumber
      libphonenumber:
        validation_method: isValidNumber
    `
		ctx := context.Background()
		ctx = makeAppContext(ctx, []byte(cfgYAML))

		f := FormatPhone{}.CheckFormat
		So(f(ctx, "+85212341234"), ShouldBeError, "invalid phone number")
		So(f(ctx, "+85298765432"), ShouldBeNil)
	})

}
