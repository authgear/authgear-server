package marshal

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestShouldNotHaveDuplicatedTypeInSamePath(t *testing.T) {
	Convey("Test ShouldNotHaveDuplicatedTypeInSamePath", t, func() {
		Convey("should pass for normal struct", func() {
			type SubConfigItem struct {
				Num1 *int `default_zero_value:"true"`
			}

			type ConfigItem struct {
				SubItem *SubConfigItem `default_zero_value:"true"`
			}

			type RootConfig struct {
				Item *ConfigItem `default_zero_value:"true"`
			}

			pathSet := map[string]interface{}{}
			pass := ShouldNotHaveDuplicatedTypeInSamePath(&RootConfig{}, pathSet)
			So(pass, ShouldBeTrue)
		})

		Convey("should fail for struct with self reference", func() {
			type ConfigItem struct {
				SubItem *ConfigItem `default_zero_value:"true"`
			}

			type RootConfig struct {
				Item *ConfigItem `default_zero_value:"true"`
			}

			pathSet := map[string]interface{}{}
			pass := ShouldNotHaveDuplicatedTypeInSamePath(&RootConfig{}, pathSet)
			So(pass, ShouldBeFalse)
		})
	})
}

func TestUpdateNilFieldsWithZeroValue(t *testing.T) {
	type ChildStruct struct {
		Num1 *int
		Num2 *int `default_zero_value:"true"`
	}

	type TestStruct struct {
		ChildNode1 *ChildStruct `default_zero_value:"true"`
		ChildNode2 *ChildStruct
	}

	Convey("UpdateNilFieldsWithZeroValue", t, func() {
		Convey("should update nil fields with tag", func() {
			s := &TestStruct{}

			UpdateNilFieldsWithZeroValue(s)
			So(s.ChildNode1, ShouldNotBeNil)
			So(s.ChildNode2, ShouldBeNil)
			So(s.ChildNode1.Num1, ShouldBeNil)
			So(s.ChildNode1.Num2, ShouldNotBeNil)
		})
	})
}

func TestOmitEmpty(t *testing.T) {
	type A struct {
		Str string
	}
	type B struct {
		A *A
	}
	Convey("OmitEmpty", t, func() {
		Convey("Map", func() {
			zero := 0
			m1 := map[string]*int{
				"a": &zero,
			}
			OmitEmpty(m1)
			So(len(m1), ShouldEqual, 0)
		})
		Convey("Slice", func() {
			slice := []*A{&A{}, &A{}}
			OmitEmpty(slice)
			So(slice[0], ShouldBeNil)
			So(slice[1], ShouldBeNil)
		})
		Convey("Struct", func() {
			b := &B{
				A: &A{},
			}
			OmitEmpty(b)
			So(b.A, ShouldBeNil)
		})
	})
}
