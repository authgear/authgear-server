package configsource

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestAuthgearYAML(t *testing.T) {
	Convey("AuthgearYAML custom attributes", t, func() {
		path := "authgear.yaml"
		featureConfig := config.NewEffectiveDefaultFeatureConfig()
		ctx := context.Background()
		ctx = context.WithValue(ctx, ContextKeyFeatureConfig, featureConfig)
		app := resource.LeveledAferoFs{FsLevel: resource.FsLevelApp}
		descriptor := &AuthgearYAMLDescriptor{}

		Convey("Custom attributes cannot be removed", func() {
			_, err := descriptor.UpdateResource(
				ctx,
				nil,
				&resource.ResourceFile{
					Location: resource.Location{
						Fs:   app,
						Path: path,
					},
					Data: []byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0000"
      pointer: /a
      type: string
`),
				},
				[]byte(`id: test
http:
  public_origin: http://test
`),
			)
			So(err, ShouldBeError, `invalid authgear.yaml:
/user_profile/custom_attributes/attributes: custom attribute of id '0000' was deleted`)
		})

		Convey("Custom attribute ID cannot be changed", func() {
			_, err := descriptor.UpdateResource(
				ctx,
				nil,
				&resource.ResourceFile{
					Location: resource.Location{
						Fs:   app,
						Path: path,
					},
					Data: []byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0000"
      pointer: /a
      type: string
`),
				},
				[]byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0001"
      pointer: /a
      type: string
`),
			)
			So(err, ShouldBeError, `invalid authgear.yaml:
/user_profile/custom_attributes/attributes: custom attribute of id '0000' was deleted`)
		})

		Convey("Custom attribute type cannot be changed", func() {
			_, err := descriptor.UpdateResource(
				ctx,
				nil,
				&resource.ResourceFile{
					Location: resource.Location{
						Fs:   app,
						Path: path,
					},
					Data: []byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0000"
      pointer: /a
      type: string
`),
				},
				[]byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0000"
      pointer: /a
      type: integer
`),
			)
			So(err, ShouldBeError, `invalid authgear.yaml:
/user_profile/custom_attributes/attributes/0: custom attribute of id '0000' has type changed; original: string, incoming: integer`)
		})

		Convey("Custom attribute can be added", func() {
			_, err := descriptor.UpdateResource(
				ctx,
				nil,
				&resource.ResourceFile{
					Location: resource.Location{
						Fs:   app,
						Path: path,
					},
					Data: []byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0000"
      pointer: /a
      type: string
`),
				},
				[]byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0000"
      pointer: /a
      type: string
    - id: "0001"
      pointer: /b
      type: string
`),
			)
			So(err, ShouldBeNil)
		})

		Convey("Custom attribute can be reordered", func() {
			_, err := descriptor.UpdateResource(
				ctx,
				nil,
				&resource.ResourceFile{
					Location: resource.Location{
						Fs:   app,
						Path: path,
					},
					Data: []byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0000"
      pointer: /a
      type: string
    - id: "0001"
      pointer: /b
      type: string
`),
				},
				[]byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0001"
      pointer: /b
      type: string
    - id: "0000"
      pointer: /a
      type: string
`),
			)
			So(err, ShouldBeNil)
		})
	})
}
