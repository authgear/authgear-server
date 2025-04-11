package cmdsetup

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/util/bubbleteautil"
)

type QuestionName int

const (
	QuestionName_AcceptAgreement QuestionName = iota

	QuestionName_EnterDomain_Project
	QuestionName_EnterDomain_Portal
	QuestionName_EnterDomain_Accounts
	QuestionName_EnableCertbot
	QuestionName_SelectCertbotEnvironment

	QuestionName_EnterAdminEmail
	QuestionName_EnterAdminPassword
	QuestionName_EnterAdminPassword_Confirm

	QuestionName_SelectSMTP
	QuestionName_EnterSMTPHost
	QuestionName_EnterSMTPPort
	QuestionName_EnterSMTPUsername
	QuestionName_EnterSMTPPassword
	QuestionName_EnterSMTPSenderAddress
	QuestionName_EnterSendgridAPIKey

	QuestionName_EnterTestEmailAddress
	QuestionName_AskForTestEmailResult
)

const (
	ValueTrue  = "true"
	ValueFalse = "false"
)

var Question_AcceptAgreement = Question{
	Name: QuestionName_AcceptAgreement,
	Model: bubbleteautil.NewSimplePicker(bubbleteautil.SimplePicker{
		Title: `License Agreement

You must accept the license terms to proceed
Authgear ONCE license agreement: https://authgear.com/once/license

`,
		Prompt: "I've read and accept the terms of Authgear ONCE license agreement",
		Items: []bubbleteautil.SimplePickerItem{
			{
				Label: "Yes",
				Value: ValueTrue,
			},
			{
				Label: "No",
				Value: ValueFalse,
			},
		},
		ValidateFunc: func(value string) error {
			if value != ValueTrue {
				return fmt.Errorf("You must accept the agreement to proceed.")
			}
			return nil
		},
	}),
}

var Question_EnterDomain_Project = Question{
	Name: QuestionName_EnterDomain_Project,
	Model: bubbleteautil.NewSingleLineTextInput(bubbleteautil.SingleLineTextInput{
		Title:        "Domain Setup\n\n",
		Prompt:       "Domain of your project (e.g. auth.mybusiness.com)",
		ValidateFunc: validateDomain,
	}),
}

var Question_EnterDomain_Portal = Question{
	Name: QuestionName_EnterDomain_Portal,
	Model: bubbleteautil.NewSingleLineTextInput(bubbleteautil.SingleLineTextInput{
		Prompt:       "Domain of Authgear portal (e.g. authgear-portal.mybusiness.com)",
		ValidateFunc: validateDomain,
	}),
}

var Question_EnterDomain_Accounts = Question{
	Name: QuestionName_EnterDomain_Accounts,
	Model: bubbleteautil.NewSingleLineTextInput(bubbleteautil.SingleLineTextInput{
		Prompt:       "Domain to login Authgear portal (e.g. authgear-portal-accounts.mybusiness.com)",
		ValidateFunc: validateDomain,
	}),
}

var Question_EnterAdminEmail = Question{
	Name: QuestionName_EnterAdminEmail,
	Model: bubbleteautil.NewSingleLineTextInput(bubbleteautil.SingleLineTextInput{
		Title:        "Admin user\n\n",
		Prompt:       "Admin email to login to Authgear portal",
		ValidateFunc: validateEmailAddress,
	}),
}

var Question_EnableCertbot = Question{
	Name: QuestionName_EnableCertbot,
	Model: bubbleteautil.NewSimplePicker(bubbleteautil.SimplePicker{
		Prompt: "Obtain TLS certificates issued by Let's Encrypt with certbot for the above domains",
		Items: []bubbleteautil.SimplePickerItem{
			{
				Label: "No, I will have TLS termination elsewhere.",
				Value: ValueFalse,
			},
			{
				Label: "Yes, I assure the DNS records of the above domains are properly set up, and the traffic is routed to this machine.",
				Value: ValueTrue,
			},
		},
	}),
}

const (
	CertbotEnvironmentProduction = "production"
	CertbotEnvironmentStaging    = "staging"
)

var Question_SelectCertbotEnvironment = Question{
	Name: QuestionName_SelectCertbotEnvironment,
	Model: bubbleteautil.NewSimplePicker(bubbleteautil.SimplePicker{
		Prompt: "Select Let's Encrypt environment",
		Items: []bubbleteautil.SimplePickerItem{
			{
				Label: "Production. This is what you typically want to use.",
				Value: CertbotEnvironmentProduction,
			},
			{
				Label: "Staging. Use this only for troubleshooting.",
				Value: CertbotEnvironmentStaging,
			},
		},
	}),
}

var Question_EnterAdminPassword = Question{
	Name: QuestionName_EnterAdminPassword,
	Model: bubbleteautil.NewSingleLineTextInput(bubbleteautil.SingleLineTextInput{
		Prompt:       "Create a password for the admin user",
		ValidateFunc: validatePassword,
		IsMasked:     true,
	}),
}

var Question_EnterAdminPassword_Confirm = Question{
	Name: QuestionName_EnterAdminPassword_Confirm,
	Model: bubbleteautil.NewSingleLineTextInput(bubbleteautil.SingleLineTextInput{
		Prompt:   "Re-enter the password to confirm",
		IsMasked: true,
	}),
}

const (
	SMTPSendgrid = "sendgrid"
	SMTPCustom   = "custom"
	SMTPSkip     = "skip"
)

var Question_SelectSMTP = Question{
	Name: QuestionName_SelectSMTP,
	Model: bubbleteautil.NewSimplePicker(bubbleteautil.SimplePicker{
		Title: `Email provider

You must configure a email provider to use Authgear. It's used for verifying email addresses and sending other system emails.

`,
		Prompt: "Email provider",
		Items: []bubbleteautil.SimplePickerItem{
			{
				Label: "Sendgrid",
				Value: SMTPSendgrid,
			},
			{
				Label: "SMTP",
				Value: SMTPCustom,
			},
			{
				Label: "Skip, set up later in the portal",
				Value: SMTPSkip,
			},
		},
	}),
}

var Question_EnterSMTPHost = Question{
	Name: QuestionName_EnterSMTPHost,
	Model: bubbleteautil.NewSingleLineTextInput(bubbleteautil.SingleLineTextInput{
		Prompt:       "SMTP server host (e.g. smtp.sendgrid.net)",
		ValidateFunc: validateDomain,
	}),
}

var Question_EnterSMTPPort = Question{
	Name: QuestionName_EnterSMTPPort,
	Model: bubbleteautil.NewSingleLineTextInput(bubbleteautil.SingleLineTextInput{
		Prompt:       "SMTP server port (commonly 25 or 587)",
		ValidateFunc: validatePort,
	}),
}

var Question_EnterSMTPUsername = Question{
	Name: QuestionName_EnterSMTPUsername,
	Model: bubbleteautil.NewSingleLineTextInput(bubbleteautil.SingleLineTextInput{
		Prompt:       "SMTP username",
		ValidateFunc: validateSMTPUsername,
	}),
}

var Question_EnterSMTPPassword = Question{
	Name: QuestionName_EnterSMTPPassword,
	Model: bubbleteautil.NewSingleLineTextInput(bubbleteautil.SingleLineTextInput{
		Prompt:       "SMTP password",
		ValidateFunc: validateSMTPPassword,
		IsMasked:     true,
	}),
}

var Question_EnterSMTPSenderAddress = Question{
	Name: QuestionName_EnterSMTPSenderAddress,
	Model: bubbleteautil.NewSingleLineTextInput(bubbleteautil.SingleLineTextInput{
		Prompt:       "Sender email address, the system emails will be sent from this email address",
		ValidateFunc: validateEmailAddress,
	}),
}

var Question_EnterSendgridAPIKey = Question{
	Name: QuestionName_EnterSendgridAPIKey,
	Model: bubbleteautil.NewSingleLineTextInput(bubbleteautil.SingleLineTextInput{
		Prompt:       "A Sendgrid API key with enough permission. See https://www.twilio.com/docs/sendgrid/for-developers/sending-email/getting-started-smtp",
		ValidateFunc: validateSendgridAPIKey,
		IsMasked:     true,
	}),
}

var Question_EnterTestEmailAddress = Question{
	Name: QuestionName_EnterTestEmailAddress,
	Model: bubbleteautil.NewSingleLineTextInput(bubbleteautil.SingleLineTextInput{
		Title:        "Testing email service\n",
		Prompt:       "Enter an email address to receive a test email",
		ValidateFunc: validateEmailAddress,
	}),
}

const (
	SendTestEmailResultSuccess                 = "success"
	SendTestEmailResultCorrectSenderAndRetry   = "correct-sender-and-retry"
	SendTestEmailResultReconfigureSMTPAndRetry = "reconfigure-smtp-and-retry"
)

var Question_AskForTestEmailResult = Question{
	Name: QuestionName_AskForTestEmailResult,
	Model: bubbleteautil.NewSimplePicker(bubbleteautil.SimplePicker{
		Prompt: "TO BE REPLACED",
		Items: []bubbleteautil.SimplePickerItem{
			{
				Label: "Yes",
				Value: SendTestEmailResultSuccess,
			},
			{
				Label: "No, retry sending email",
				Value: SendTestEmailResultCorrectSenderAndRetry,
			},
			{
				Label: "No, review email provider setup",
				Value: SendTestEmailResultReconfigureSMTPAndRetry,
			},
		},
	}),
}
