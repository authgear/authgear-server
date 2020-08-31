package template

import (
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEngineRender(t *testing.T) {
	Convey("Engine.Render", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resolver := NewMockTemplateResolver(ctrl)
		engine := &Engine{Resolver: resolver}

		resolver.EXPECT().Resolve(gomock.Any(), gomock.Any()).Return(&Resolved{
			T: T{
				Type:   "page-a",
				IsHTML: true,
				Defines: []string{
					`
					{{- define "define-a" -}}
					define-a
					{{- end -}}
					`,
				},
			},
			Content: `<!DOCTYPE html>
<html>
<head><title>Hi</title></head>
<body>
{{ template "define-a" }}
{{ template "component-a" }}
<p>{{ template "greeting" (makemap "URL" .URL) }}</p>
</body>
</html>`,
			Translations: map[string]Translation{
				"greeting": Translation{
					LanguageTag: "en",
					Value:       `<a href="{URL}">Hi</a>`,
				},
			},
			ComponentContents: []string{
				`
				{{- define "component-a" -}}
				component-a
				{{- end -}}
				`,
			},
		}, nil)

		out, err := engine.Render(&RenderContext{
			ValidatorOptions: []ValidatorOption{
				AllowRangeNode(true),
				AllowTemplateNode(true),
				AllowDeclaration(true),
				MaxDepth(15),
			},
		}, "", map[string]interface{}{
			"URL": "http://www.example.com",
		})
		So(err, ShouldBeNil)
		So(out, ShouldEqual, `<!DOCTYPE html>
<html>
<head><title>Hi</title></head>
<body>
define-a
component-a
<p><a href="http://www.example.com">Hi</a></p>
</body>
</html>`)
	})
}

func TestEngineRenderTranslation(t *testing.T) {
	Convey("Engine.RenderTranslation", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resolver := NewMockTemplateResolver(ctrl)
		engine := &Engine{Resolver: resolver}

		resolver.EXPECT().ResolveTranslations(gomock.Any(), "translation.json").Return(map[string]Translation{
			`"greeting"`: {
				LanguageTag: "en",
				Value:       `<a href="{URL}">Hi</a>`,
			},
		}, nil)

		out, err := engine.RenderTranslation(&RenderContext{
			ValidatorOptions: []ValidatorOption{
				AllowRangeNode(true),
				AllowTemplateNode(true),
				AllowDeclaration(true),
				MaxDepth(15),
			},
		}, "translation.json", `"greeting"`, map[string]interface{}{
			"URL": "http://www.example.com",
		})
		So(err, ShouldBeNil)
		So(out, ShouldEqual, `<a href="http://www.example.com">Hi</a>`)
	})
}
