package template_test

import (
	"context"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func TestTemplateResource(t *testing.T) {
	Convey("HTML EffectiveResource", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelCustom},
		})

		txt := &template.HTML{Name: "resource.txt"}
		r.Register(txt)

		writeFile := func(fs afero.Fs, lang string, data string) {
			_ = fs.MkdirAll("templates/"+lang, 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/resource.txt", []byte(data), 0666)
		}

		read := func(view resource.View) (str string, err error) {
			result, err := manager.Read(txt, view)
			if err != nil {
				return
			}

			h := result.(*template.HTMLTemplateEffectiveResource)
			tpl := h.Template
			var out strings.Builder
			err = tpl.Execute(&out, nil)
			if err != nil {
				return
			}

			str = out.String()
			return
		}

		Convey("it should return single resource", func() {
			writeFile(fsA, "en", "en in fs A")

			data, err := read(resource.EffectiveResource{
				DefaultTag:    "en",
				SupportedTags: []string{"en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")
		})

		Convey("it should return resource with preferred language", func() {
			writeFile(fsA, "en", "en in fs A")
			writeFile(fsA, "zh", "zh in fs A")

			data, err := read(resource.EffectiveResource{
				DefaultTag:    "en",
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")

			data, err = read(resource.EffectiveResource{
				DefaultTag:    "en",
				PreferredTags: []string{"en"},
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")

			data, err = read(resource.EffectiveResource{
				DefaultTag:    "en",
				PreferredTags: []string{"zh"},
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "zh in fs A")

		})

		Convey("it should return something if default language is not found", func() {
			writeFile(fsA, "en", "en in fs A")
			writeFile(fsA, "zh", "zh in fs A")

			data, err := read(resource.EffectiveResource{
				DefaultTag:    "en",
				PreferredTags: []string{"ja"},
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")
		})

		Convey("it should combine resources in different FS", func() {
			writeFile(fsB, "zh", "zh in fs B")
			writeFile(fsA, "en", "en in fs A")

			data, err := read(resource.EffectiveResource{
				DefaultTag:    "en",
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")

			data, err = read(resource.EffectiveResource{
				DefaultTag:    "en",
				PreferredTags: []string{"en"},
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")

			data, err = read(resource.EffectiveResource{
				DefaultTag:    "en",
				PreferredTags: []string{"zh"},
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "zh in fs B")
		})

		Convey("it should not fail when fallback is not en", func() {
			writeFile(fsA, "en", "en in fs A")

			data, err := read(resource.EffectiveResource{
				DefaultTag:    "zh",
				SupportedTags: []string{"zh"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")
		})
	})

	Convey("HTML EffectiveFile", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelCustom},
		})

		txt := &template.HTML{Name: "resource.txt"}
		r.Register(txt)

		writeFile := func(fs afero.Fs, lang string, data string) {
			_ = fs.MkdirAll("templates/"+lang, 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/resource.txt", []byte(data), 0666)
		}

		read := func(lang string) (str string, err error) {
			view := resource.EffectiveFile{
				Path: "templates/" + lang + "/resource.txt",
			}
			result, err := manager.Read(txt, view)
			if err != nil {
				return
			}

			str = string(result.([]byte))
			return
		}

		Convey("it should return single resource", func() {
			writeFile(fsA, "en", "en in fs A")

			data, err := read("en")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")

			_, err = read("zh")
			So(err, ShouldBeError, "specified resource is not configured")
		})

		Convey("it should return resource with preferred language", func() {
			writeFile(fsA, "en", "en in fs A")
			writeFile(fsA, "zh", "zh in fs A")

			data, err := read("en")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")

			data, err = read("zh")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "zh in fs A")
		})

		Convey("it should combine resources in different FS", func() {
			writeFile(fsB, "zh", "zh in fs B")
			writeFile(fsA, "en", "en in fs A")

			data, err := read("en")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")

			data, err = read("zh")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "zh in fs B")
		})
	})

	Convey("HTML AppFile", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelApp},
		})

		txt := &template.MessageHTML{Name: "messages/resource.txt"}
		r.Register(txt)

		writeFile := func(fs afero.Fs, lang string, data string) {
			_ = fs.MkdirAll("templates/"+lang+"/messages", 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/messages/resource.txt", []byte(data), 0666)
		}

		read := func(lang string) (str string, err error) {
			view := resource.AppFile{
				Path: "templates/" + lang + "/messages/resource.txt",
			}
			result, err := manager.Read(txt, view)
			if err != nil {
				return
			}

			str = string(result.([]byte))
			return
		}

		Convey("not found", func() {
			writeFile(fsA, "en", "en in fs A")

			_, err := read("en")
			So(err, ShouldBeError, "specified resource is not configured")
		})

		Convey("found", func() {
			writeFile(fsB, "en", "en in fs B")

			data, err := read("en")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs B")
		})

		Convey("it should return resource in app FS", func() {
			writeFile(fsA, "en", "en in fs A")
			writeFile(fsB, "en", "en in fs B")

			data, err := read("en")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs B")
		})
	})

	Convey("matchTemplatePath", t, func() {
		// expected path "templates/{{ localeKey }}/.../{{ templateName }}"
		mockRes := &template.MessageHTML{
			Name: "messages/dummy.html",
		}

		Convey("path with less than 3 segments", func() {
			_, result := mockRes.MatchResource(
				"templates/dummy.html",
			)
			So(result, ShouldBeFalse)
		})

		Convey("first segment mismatch", func() {
			_, result := mockRes.MatchResource(
				"wrong/en/messages/dummy.html",
			)
			So(result, ShouldBeFalse)
		})

		Convey("invalid locale key", func() {
			_, result := mockRes.MatchResource(
				"templates/abc/messages/dummy.html",
			)
			So(result, ShouldBeFalse)
		})

		Convey("template name mismatch", func() {
			_, result := mockRes.MatchResource(
				"templates/en/messages/wrong_name.html",
			)
			So(result, ShouldBeFalse)
		})

		Convey("valid path", func() {
			match, result := mockRes.MatchResource(
				"templates/en/messages/dummy.html",
			)
			So(result, ShouldBeTrue)
			So(match.LanguageTag, ShouldEqual, "en")
		})
	})

	Convey("Match html vs message html", t, func() {
		// expected path "templates/{{ localeKey }}/.../{{ templateName }}"
		html := &template.HTML{
			Name: "dummy.html",
		}
		messageHtml := &template.MessageHTML{
			Name: "messages/dummy.html",
		}
		Convey("html should not match messages/ prefix", func() {
			_, messageRes := html.MatchResource("templates/en/messages/dummy.html")
			So(messageRes, ShouldBeFalse)
			_, rootRes := html.MatchResource("templates/en/dummy.html")
			So(rootRes, ShouldBeTrue)
			_, folderRes := html.MatchResource("templates/en/other/dummy.html")
			So(folderRes, ShouldBeFalse)
			_, nestedFolderRes := html.MatchResource("templates/en/other/messages/dummy.html")
			So(nestedFolderRes, ShouldBeFalse)
		})
		Convey("message html should only match messages/ prefix", func() {
			_, messageRes := messageHtml.MatchResource("templates/en/messages/dummy.html")
			So(messageRes, ShouldBeTrue)
			_, rootRes := messageHtml.MatchResource("templates/en/dummy.html")
			So(rootRes, ShouldBeFalse)
			_, folderRes := messageHtml.MatchResource("templates/en/other/dummy.html")
			So(folderRes, ShouldBeFalse)
			_, nestedFolderRes := messageHtml.MatchResource("templates/en/other/messages/dummy.html")
			So(nestedFolderRes, ShouldBeFalse)
		})
	})

	Convey("Match txt vs message txt", t, func() {
		// expected path "templates/{{ localeKey }}/.../{{ templateName }}"
		txt := &template.PlainText{
			Name: "dummy.txt",
		}
		messageTxt := &template.MessagePlainText{
			Name: "messages/dummy.txt",
		}
		Convey("txt should not match messages/ prefix", func() {
			_, messageRes := txt.MatchResource("templates/en/messages/dummy.txt")
			So(messageRes, ShouldBeFalse)
			_, rootRes := txt.MatchResource("templates/en/dummy.txt")
			So(rootRes, ShouldBeTrue)
			_, folderRes := txt.MatchResource("templates/en/other/dummy.txt")
			So(folderRes, ShouldBeFalse)
			_, nestedFolderRes := txt.MatchResource("templates/en/other/messages/dummy.txt")
			So(nestedFolderRes, ShouldBeFalse)
		})
		Convey("message txt should only match messages/ prefix", func() {
			_, messageRes := messageTxt.MatchResource("templates/en/messages/dummy.txt")
			So(messageRes, ShouldBeTrue)
			_, rootRes := messageTxt.MatchResource("templates/en/dummy.txt")
			So(rootRes, ShouldBeFalse)
			_, folderRes := messageTxt.MatchResource("templates/en/other/dummy.txt")
			So(folderRes, ShouldBeFalse)
			_, nestedFolderRes := messageTxt.MatchResource("templates/en/other/messages/dummy.txt")
			So(nestedFolderRes, ShouldBeFalse)
		})
	})

	Convey("Feature flag for update html resource", t, func() {
		app := resource.LeveledAferoFs{FsLevel: resource.FsLevelApp}
		path := "templates/en/messages/dummy.html"
		messageHtml := &template.MessageHTML{
			Name: "messages/dummy.html",
		}
		resourceFile := resource.ResourceFile{
			Location: resource.Location{
				Fs:   app,
				Path: path,
			},
			Data: []byte("qwer"),
		}
		Convey("Should not allow update if disallowed", func() {
			featureConfig := config.NewEffectiveDefaultFeatureConfig()
			featureConfig.Messaging.TemplateCustomizationDisabled = true
			ctx := context.Background()
			ctx = context.WithValue(ctx, configsource.ContextKeyFeatureConfig, featureConfig)
			res, err := messageHtml.UpdateResource(ctx, nil, &resourceFile, []byte("asdf"))
			So(res, ShouldBeNil)
			So(err, ShouldEqual, template.ErrUpdateDisallowed)
		})
		Convey("Should allow update if flag not set", func() {
			featureConfig := config.NewEffectiveDefaultFeatureConfig()
			ctx := context.Background()
			ctx = context.WithValue(ctx, configsource.ContextKeyFeatureConfig, featureConfig)
			res, err := messageHtml.UpdateResource(ctx, nil, &resourceFile, []byte("asdf"))
			So(res, ShouldNotBeNil)
			So(res.Data, ShouldEqual, []byte("asdf"))
			So(err, ShouldBeNil)
		})
	})

	Convey("Feature flag for update txt resource", t, func() {
		app := resource.LeveledAferoFs{FsLevel: resource.FsLevelApp}
		path := "templates/en/messages/dummy.txt"
		messageTxt := &template.MessagePlainText{
			Name: "messages/dummy.txt",
		}
		resourceFile := resource.ResourceFile{
			Location: resource.Location{
				Fs:   app,
				Path: path,
			},
			Data: []byte("qwer"),
		}
		Convey("Should not allow update if disallowed", func() {
			featureConfig := config.NewEffectiveDefaultFeatureConfig()
			featureConfig.Messaging.TemplateCustomizationDisabled = true
			ctx := context.Background()
			ctx = context.WithValue(ctx, configsource.ContextKeyFeatureConfig, featureConfig)
			res, err := messageTxt.UpdateResource(ctx, nil, &resourceFile, []byte("asdf"))
			So(res, ShouldBeNil)
			So(err, ShouldEqual, template.ErrUpdateDisallowed)
		})
		Convey("Should allow update if flag not set", func() {
			featureConfig := config.NewEffectiveDefaultFeatureConfig()
			ctx := context.Background()
			ctx = context.WithValue(ctx, configsource.ContextKeyFeatureConfig, featureConfig)
			res, err := messageTxt.UpdateResource(ctx, nil, &resourceFile, []byte("asdf"))
			So(res, ShouldNotBeNil)
			So(res.Data, ShouldEqual, []byte("asdf"))
			So(err, ShouldBeNil)
		})
	})

	Convey("Template validation on view validate resource", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelCustom},
		})

		txt := &template.MessagePlainText{Name: "resource.txt"}
		html := &template.MessageHTML{Name: "resource.html"}
		r.Register(txt)
		r.Register(html)

		writeFileTxt := func(fs afero.Fs, lang string, data string) {
			_ = fs.MkdirAll("templates/"+lang, 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/resource.txt", []byte(data), 0666)
		}
		writeFileHTML := func(fs afero.Fs, lang string, data string) {
			_ = fs.MkdirAll("templates/"+lang, 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/resource.html", []byte(data), 0666)
		}

		readTxtAndValidate := func(view resource.ValidateResourceView) (str string, err error) {
			_, err = manager.Read(txt, view)
			if err != nil {
				return
			}
			return
		}
		readHTMLAndValidate := func(view resource.ValidateResourceView) (str string, err error) {
			_, err = manager.Read(html, view)
			if err != nil {
				return
			}
			return
		}

		templateStrOf100JSCalls := `{{ template "name" (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js (js "\\")))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))))) }}`
		Convey("Should throw err if invalid message text template", func() {
			writeFileTxt(fsA, "en", templateStrOf100JSCalls)

			_, err := readTxtAndValidate(resource.ValidateResource{})
			So(err, ShouldBeError, "invalid text template: resource.txt:1:408: template nested too deep")
		})

		Convey("Should run normally if valid message text template", func() {
			writeFileTxt(fsA, "en", `{{ template "name" }}`)

			_, err := readTxtAndValidate(resource.ValidateResource{})
			So(err, ShouldBeNil)

		})

		Convey("Should throw err if invalid message html template", func() {
			writeFileHTML(fsA, "en", templateStrOf100JSCalls)

			_, err := readHTMLAndValidate(resource.ValidateResource{})
			So(err, ShouldBeError, "invalid HTML template: resource.html:1:408: template nested too deep")
		})

		Convey("Should run normally if valid message html template", func() {
			writeFileHTML(fsA, "en", `{{ template "name" }}`)

			_, err := readHTMLAndValidate(resource.ValidateResource{})
			So(err, ShouldBeNil)
		})
	})
}
