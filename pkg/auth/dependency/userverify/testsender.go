package userverify

import (
	"fmt"

	"github.com/go-gomail/gomail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	authTemplate "github.com/skygeario/skygear-server/pkg/auth/template"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type TestCodeSender interface {
	Send(recordKey string, recordValue string) error
}

type TestEmailCodeSenderConfig struct {
	Sender      string
	SenderName  string
	Subject     string
	ReplyTo     string
	ReplyToName string
}

type TestEmailCodeSender struct {
	AppName        string
	Config         TestEmailCodeSenderConfig
	URLPrefix      string
	Dialer         *gomail.Dialer
	TemplateEngine *template.Engine
}

func (t *TestEmailCodeSender) Send(recordKey string, recordValue string) (err error) {
	context := prepareVerifyTestRequestContext(
		recordKey,
		recordValue,
		t.AppName,
		t.URLPrefix,
	)

	var textBody string
	if textBody, err = t.TemplateEngine.ParseTextTemplate(
		authTemplate.VerifyTextTemplateNameForKey(recordKey),
		context,
		template.ParseOption{Required: true, DefaultTemplateName: authTemplate.TemplateNameVerifyEmailText},
	); err != nil {
		return
	}

	var htmlBody string
	if htmlBody, err = t.TemplateEngine.ParseTextTemplate(
		authTemplate.VerifyHTMLTemplateNameForKey(recordKey),
		context,
		template.ParseOption{Required: false, DefaultTemplateName: authTemplate.TemplateNameVerifyEmailHTML},
	); err != nil {
		return
	}

	sendReq := mail.SendRequest{
		Dialer:      t.Dialer,
		Sender:      t.Config.Sender,
		SenderName:  t.Config.SenderName,
		Recipient:   recordValue,
		Subject:     t.Config.Subject,
		ReplyTo:     t.Config.ReplyTo,
		ReplyToName: t.Config.ReplyToName,
		TextBody:    textBody,
		HTMLBody:    htmlBody,
	}

	err = sendReq.Execute()
	return
}

type TestSMSCodeSender struct {
	AppName        string
	URLPrefix      string
	SMSClient      sms.Client
	TemplateEngine *template.Engine
}

func (t *TestSMSCodeSender) Send(recordKey string, recordValue string) (err error) {
	context := prepareVerifyTestRequestContext(
		recordKey,
		recordValue,
		t.AppName,
		t.URLPrefix,
	)

	var textBody string
	if textBody, err = t.TemplateEngine.ParseTextTemplate(
		authTemplate.VerifyTextTemplateNameForKey(recordKey),
		context,
		template.ParseOption{Required: true, DefaultTemplateName: authTemplate.TemplateNameVerifySMSText},
	); err != nil {
		return
	}

	err = t.SMSClient.Send(recordValue, textBody)
	return
}

func prepareVerifyTestRequestContext(
	recordKey string,
	recordValue string,
	appName string,
	urlPrefix string,
) map[string]interface{} {
	userProfile := userprofile.UserProfile{
		Meta: userprofile.Meta{
			ID:         "user/dummy-id",
			RecordID:   "dummy-id",
			RecordType: "user",
			OwnerID:    "dummy-id",
			CreatedBy:  "dummy-id",
			UpdatedBy:  "dummy-id",
		},
		Data: userprofile.Data{},
	}
	userProfile.Data[recordKey] = recordValue
	code := "testing-code"

	return map[string]interface{}{
		"appname":      appName,
		"record_key":   recordKey,
		"record_value": recordValue,
		"user_id":      "dummy-id",
		"user":         userProfile.ToMap(),
		"code":         code,
		"link": fmt.Sprintf(
			"%s/verify_code_form?code=%s&user_id=%s",
			urlPrefix,
			code,
			userProfile.RecordID,
		),
	}
}
