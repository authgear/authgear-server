package translation_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func TestService(t *testing.T) {
	Convey("Service", t, func() {
		ctl := gomock.NewController(t)
		defer ctl.Finish()

		var TemplateMessageSMSTXT = template.RegisterMessagePlainText("messages/sms.txt")
		var TemplateMessageEmailTXT = template.RegisterMessagePlainText("messages/email.txt")
		var TemplateMessageEmailHTML = template.RegisterMessageHTML("messages/email.html")
		var TemplateMessageWhatsappTXT = template.RegisterMessagePlainText("messages/whatsapp.txt")

		var messageSpec = &translation.MessageSpec{
			MessageType:       translation.MessageTypeSetupPrimaryOOB,
			Name:              translation.SpecNameSetupPrimaryOOB,
			TXTEmailTemplate:  TemplateMessageEmailTXT,
			HTMLEmailTemplate: TemplateMessageEmailHTML,
			SMSTemplate:       TemplateMessageSMSTXT,
			WhatsappTemplate:  TemplateMessageWhatsappTXT,
		}

		ctx := context.Background()
		ctx = uiparam.WithUIParam(ctx, &uiparam.T{
			ClientID: "my client id",
			Prompt: []string{
				"my prompt",
			},
			State:     "my state",
			XState:    "my x state",
			UILocales: "my ui locales",
		})
		ctx = intl.WithPreferredLanguageTags(ctx, []string{"zh", "en"})

		fs := afero.NewMemMapFs()

		writeFile := func(lang string, name string, data string) {
			_ = fs.MkdirAll("templates/"+lang, 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/"+name, []byte(data), 0666)
		}

		// Write template.TranslationMap
		writeFile("en", "translation.json", `{
	"app.name": "My App Name",
	"email.default.sender":"no-reply@authgear.com",
	"email.default.reply-to": "",
	"email.default.subject": "",
	"email.setup-primary-oob.subject": "[{AppName}] Test",
	"sms.default.sender": "Sender: [{AppName}]"
}`)
		for _, lang := range []string{"zh", "en"} {
			for _, path := range []string{
				"messages/sms.txt",
				"messages/email.txt",
				"messages/email.html",
				"messages/whatsapp.txt",
			} {
				writeFile(lang, path, fmt.Sprintf(`%v/%v
AppName: {{ .AppName }}
ClientID: {{ .ClientID }}
Code: {{ .Code }}
Email: {{ .Email }}
HasPassword: {{ .HasPassword }}
Host: {{ .Host }}
Link: {{ .Link }}
Password: {{ .Password }}
Phone: {{ .Phone }}
State: {{ .State }}
UILocales: {{ .UILocales }}
URL: {{ .URL }}
XState: {{ .XState }}`, lang, path))
			}
		}

		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{resource.LeveledAferoFs{
			Fs:      fs,
			FsLevel: resource.FsLevelBuiltin,
		}})
		resolver := &template.Resolver{
			Resources:             manager,
			DefaultLanguageTag:    "en",
			SupportedLanguageTags: []string{"zh", "en"},
		}
		engine := &template.Engine{Resolver: resolver}

		service := translation.Service{
			Context:        ctx,
			TemplateEngine: engine,
			StaticAssets:   NewMockStaticAssetResolver(ctl),
		}

		Convey("it should render otp messages correctly", func() {
			emailMessageData, err := service.EmailMessageData(messageSpec, &translation.PartialTemplateVariables{
				Email:       "my-email@example.com",
				Phone:       "+85298765432",
				Code:        "123456",
				URL:         "https://www.example.com/url",
				Host:        "https://www.example.com",
				Link:        "https://www.example.com/link",
				HasPassword: true,
			})
			So(err, ShouldBeNil)
			So(emailMessageData.Sender, ShouldEqual, "no-reply@authgear.com")
			So(emailMessageData.ReplyTo, ShouldEqual, "")
			So(emailMessageData.Subject, ShouldEqual, "[My App Name] Test")
			So(emailMessageData.TextBody.LanguageTag, ShouldEqual, "zh")
			So(emailMessageData.TextBody.String, ShouldEqual, `zh/messages/email.txt
AppName: My App Name
ClientID: my+client+id
Code: 123456
Email: my-email@example.com
HasPassword: true
Host: https://www.example.com
Link: https://www.example.com/link
Password: 
Phone: +85298765432
State: my+state
UILocales: my+ui+locales
URL: https://www.example.com/url
XState: my+x+state`)
			So(emailMessageData.HTMLBody.LanguageTag, ShouldEqual, "zh")
			So(emailMessageData.HTMLBody.String, ShouldEqual, `zh/messages/email.html
AppName: My App Name
ClientID: my client id
Code: 123456
Email: my-email@example.com
HasPassword: true
Host: https://www.example.com
Link: https://www.example.com/link
Password: 
Phone: &#43;85298765432
State: my state
UILocales: my ui locales
URL: https://www.example.com/url
XState: my x state`)

			smsMessageData, err := service.SMSMessageData(messageSpec, &translation.PartialTemplateVariables{
				Email: "my-email@example.com",
				Phone: "+85298765432",
				Code:  "123456",
				URL:   "https://www.example.com/url",
				Host:  "https://www.example.com",
				Link:  "https://www.example.com/link",
			})
			So(err, ShouldBeNil)
			So(smsMessageData.Sender, ShouldEqual, "Sender: [My App Name]")
			So(smsMessageData.Body.LanguageTag, ShouldEqual, "zh")
			So(smsMessageData.Body.String, ShouldEqual, `zh/messages/sms.txt
AppName: My App Name
ClientID: my+client+id
Code: 123456
Email: my-email@example.com
HasPassword: false
Host: https://www.example.com
Link: https://www.example.com/link
Password: 
Phone: +85298765432
State: my+state
UILocales: my+ui+locales
URL: https://www.example.com/url
XState: my+x+state`)

			whatsappMessageData, err := service.WhatsappMessageData("en", messageSpec, &translation.PartialTemplateVariables{
				Email: "my-email@example.com",
				Phone: "+85298765432",
				Code:  "123456",
				URL:   "https://www.example.com/url",
				Host:  "https://www.example.com",
				Link:  "https://www.example.com/link",
			})
			So(err, ShouldBeNil)
			So(whatsappMessageData.Body.LanguageTag, ShouldEqual, "en")
			So(whatsappMessageData.Body.String, ShouldEqual, `en/messages/whatsapp.txt
AppName: My App Name
ClientID: my client id
Code: 123456
Email: my-email@example.com
HasPassword: false
Host: https://www.example.com
Link: https://www.example.com/link
Password: 
Phone: +85298765432
State: my state
UILocales: my ui locales
URL: https://www.example.com/url
XState: my x state`)
		})

		Convey("it should render forgot password messages correctly", func() {
			emailMessageData, err := service.EmailMessageData(messageSpec, &translation.PartialTemplateVariables{
				Email:    "email@example.com",
				Password: "P@ssw0rd",
			})
			So(err, ShouldBeNil)
			So(emailMessageData.Sender, ShouldEqual, "no-reply@authgear.com")
			So(emailMessageData.ReplyTo, ShouldEqual, "")
			So(emailMessageData.Subject, ShouldEqual, "[My App Name] Test")
			So(emailMessageData.TextBody.LanguageTag, ShouldEqual, "zh")
			So(emailMessageData.TextBody.String, ShouldEqual, `zh/messages/email.txt
AppName: My App Name
ClientID: my+client+id
Code: 
Email: email@example.com
HasPassword: false
Host: 
Link: 
Password: P@ssw0rd
Phone: 
State: my+state
UILocales: my+ui+locales
URL: 
XState: my+x+state`)
			So(emailMessageData.HTMLBody.LanguageTag, ShouldEqual, "zh")
			So(emailMessageData.HTMLBody.String, ShouldEqual, `zh/messages/email.html
AppName: My App Name
ClientID: my client id
Code: 
Email: email@example.com
HasPassword: false
Host: 
Link: 
Password: P@ssw0rd
Phone: 
State: my state
UILocales: my ui locales
URL: 
XState: my x state`)

			smsMessageData, err := service.SMSMessageData(messageSpec, &translation.PartialTemplateVariables{
				Email:    "email@example.com",
				Password: "P@ssw0rd",
			})
			So(err, ShouldBeNil)
			So(smsMessageData.Sender, ShouldEqual, "Sender: [My App Name]")
			So(smsMessageData.Body.LanguageTag, ShouldEqual, "zh")
			So(smsMessageData.Body.String, ShouldEqual, `zh/messages/sms.txt
AppName: My App Name
ClientID: my+client+id
Code: 
Email: email@example.com
HasPassword: false
Host: 
Link: 
Password: P@ssw0rd
Phone: 
State: my+state
UILocales: my+ui+locales
URL: 
XState: my+x+state`)

			whatsappMessageData, err := service.WhatsappMessageData("en", messageSpec, &translation.PartialTemplateVariables{
				Email:    "email@example.com",
				Password: "P@ssw0rd",
			})
			So(err, ShouldBeNil)
			So(whatsappMessageData.Body.LanguageTag, ShouldEqual, "en")
			So(whatsappMessageData.Body.String, ShouldEqual, `en/messages/whatsapp.txt
AppName: My App Name
ClientID: my client id
Code: 
Email: email@example.com
HasPassword: false
Host: 
Link: 
Password: P@ssw0rd
Phone: 
State: my state
UILocales: my ui locales
URL: 
XState: my x state`)
		})
	})
}
