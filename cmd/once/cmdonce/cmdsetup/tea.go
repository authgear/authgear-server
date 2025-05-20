package cmdsetup

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/authgear/authgear-server/cmd/once/cmdonce/internal"
	"github.com/authgear/authgear-server/pkg/util/bubbleteautil"
)

type Question struct {
	Name  QuestionName
	Model bubbleteautil.Model
}

var _ tea.Model = Question{}

func (q Question) Init() tea.Cmd {
	return func() tea.Msg {
		return nil
	}
}

func (q Question) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	teaModel, cmd := q.Model.Update(msg)
	q.Model = teaModel.(bubbleteautil.Model)
	return q, cmd
}

func (q Question) View() string {
	return q.Model.View()
}

func (q Question) Value() string {
	return q.Model.Value()
}

func (q Question) WithValue(val string) Question {
	q.Model = q.Model.WithValue(val)
	return q
}

type RetainedValues struct {
	SMTPHost          string
	SMTPPort          string
	SMTPUsername      string
	SMTPSenderAddress string
	TestEmailAddress  string
}

type SetupApp struct {
	Context        context.Context
	HTTPClient     *http.Client
	LicenseOptions internal.LicenseOptions

	HTTPScheme string
	IsResetup  bool

	AUTHGEAR_CERTBOT_ENABLED     bool
	AUTHGEAR_CERTBOT_ENVIRONMENT string

	QuestionName_EnableCertbot_PromptEnabled            bool
	QuestionName_SelectCertbotEnvironment_PromptEnabled bool

	AUTHGEAR_ONCE_IMAGE               string
	AUTHGEAR_ONCE_LICENSE_KEY         string
	AUTHGEAR_ONCE_MACHINE_FINGERPRINT string

	Questions      []Question
	retainedValues RetainedValues

	Loading        bool
	LoadingMessage string
	Spinner        spinner.Model

	FatalError internal.FatalError

	RecoverableErr    error
	RecoverableErrCmd tea.Cmd

	Installation *Installation
	Resetup      *Resetup
}

var _ tea.Model = SetupApp{}

type msgSetupAppInit struct{}

func SetupAppInit() tea.Msg {
	return msgSetupAppInit{}
}

type msgSetupAppStartSurvey struct{}

func SetupAppStartSurvey() tea.Msg {
	return msgSetupAppStartSurvey{}
}

type msgSendTestEmail struct {
	Opts SendTestEmailOptions
	Err  error
}

type msgAskForSendTestEmailResult struct {
	Opts SendTestEmailOptions
}

type msgSetupAppAbort struct{}

func SetupAppAbort() tea.Msg {
	return msgSetupAppAbort{}
}

type msgSetupAppEndSurvey struct{}

func SetupAppEndSurvey() tea.Msg {
	return msgSetupAppEndSurvey{}
}

type msgSetupAppInitResetup struct{}
type msgSetupAppStartResetup struct{}

type msgSetupAppActivateLicense struct{}

func SetupAppActivateLicense() tea.Msg {
	return msgSetupAppActivateLicense{}
}

type msgSetupAppActivateLicenseResult struct {
	LicenseObject *internal.LicenseObject
	Err           error
}

type msgSetupStartInstallation struct{}

func SetupAppStartInstallation() tea.Msg {
	return msgSetupStartInstallation{}
}

func (m SetupApp) Init() tea.Cmd {
	return SetupAppInit
}

func (m SetupApp) msgSetupAppInit() (tea.Model, tea.Cmd) {
	return m, SetupAppStartSurvey
}

func (m SetupApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case msgSetupAppAbort:
		return m, tea.Quit
	case msgSetupAppInit:
		return m.msgSetupAppInit()
	case msgSetupAppStartSurvey:
		var cmd tea.Cmd
		m, cmd = m.appendNextQuestion()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, SetupAppAbort
		case tea.KeyEnter:
			if m.RecoverableErr != nil {
				m.RecoverableErr = nil
				if m.RecoverableErrCmd != nil {
					cmds = append(cmds, m.RecoverableErrCmd)
					m.RecoverableErrCmd = nil
				}
			} else if m.IsCurrentQuestionFocused() {
				var cmd tea.Cmd
				m, cmd = m.appendNextQuestion()
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		default:
			// Hide error in case the user hit other keys.
			cmds = append(cmds, bubbleteautil.HideError)
		}
	case msgSendTestEmail:
		m.StopLoading()
		errRecoverCmd := func() tea.Msg {
			return msgAskForSendTestEmailResult{Opts: msg.Opts}
		}
		if msg.Err != nil {
			m.RecoverableErr = msg.Err
			m.RecoverableErrCmd = errRecoverCmd
		} else {
			cmds = append(cmds, errRecoverCmd)
		}
	case msgAskForSendTestEmailResult:
		q := newQuestion(Question_AskForTestEmailResult)
		picker := q.Model.(bubbleteautil.SimplePicker)
		picker.Prompt = fmt.Sprintf("A test email is sent to %v, did you receive it", msg.Opts.ToAddress)
		q.Model = picker
		m.Questions = append(m.Questions, q)
	case msgSetupAppEndSurvey:
		m = m.blurCurrentQuestion()
		return m, SetupAppActivateLicense
	case msgSetupAppActivateLicense:
		return m, tea.Batch(
			m.StartLoading("Activating license..."),
			func() tea.Msg {
				licenseObject, err := internal.ActivateLicense(m.Context, m.HTTPClient, m.LicenseOptions)
				return msgSetupAppActivateLicenseResult{
					LicenseObject: licenseObject,
					Err:           err,
				}
			},
		)
	case msgSetupAppActivateLicenseResult:
		m.StopLoading()
		errRecoverCmd := SetupAppActivateLicense
		if msg.Err != nil {
			m.RecoverableErr = msg.Err
			m.RecoverableErrCmd = errRecoverCmd
		} else {
			installation := m.ToInstallation(msg.LicenseObject)
			m.Installation = &installation
			return m, SetupAppStartInstallation
		}
	case msgSetupAppInitResetup:
		m = m.blurCurrentQuestion()
		resetup := m.ToResetup(m.LicenseOptions)
		m.Resetup = &resetup
		return m, func() tea.Msg { return msgSetupAppStartResetup{} }
	}

	for idx := range m.Questions {
		var updated tea.Model
		var newCmd tea.Cmd
		updated, newCmd = m.Questions[idx].Update(msg)
		m.Questions[idx] = updated.(Question)
		cmds = append(cmds, newCmd)
	}
	var updated spinner.Model
	var newCmd tea.Cmd
	updated, newCmd = m.Spinner.Update(msg)
	m.Spinner = updated
	cmds = append(cmds, newCmd)

	if m.Installation != nil {
		updated, newCmd := m.Installation.Update(msg)
		installation := updated.(Installation)
		m.Installation = &installation
		cmds = append(cmds, newCmd)
	}
	if m.Resetup != nil {
		updated, newCmd := m.Resetup.Update(msg)
		resetup := updated.(Resetup)
		m.Resetup = &resetup
		cmds = append(cmds, newCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m SetupApp) performSimpleValidation() (SetupApp, bool) {
	updated, valid := m.Questions[len(m.Questions)-1].Model.Validate()
	m.Questions[len(m.Questions)-1].Model = updated.(bubbleteautil.Model)
	if !valid {
		return m, true
	}
	return m, false
}

func (m SetupApp) performCrossFieldValidation() (SetupApp, bool) {
	switch m.Questions[len(m.Questions)-1].Name {
	case QuestionName_EnterDomain_Project:
		// This is the first domain entered.
		// So it cannot be equal to any previous values.
	case QuestionName_EnterDomain_Portal:
		project := m.mustFindQuestionByName(QuestionName_EnterDomain_Project).Value()
		portal := m.mustFindQuestionByName(QuestionName_EnterDomain_Portal).Value()

		if portal == project {
			err := fmt.Errorf("It cannot be equal to %v. Please enter a different value", project)
			m.Questions[len(m.Questions)-1].Model = m.Questions[len(m.Questions)-1].Model.WithError(err)
			return m, true
		} else {
			m.Questions[len(m.Questions)-1].Model = m.Questions[len(m.Questions)-1].Model.WithError(nil)
		}
	case QuestionName_EnterDomain_Accounts:
		// The is the last domain entered.
		project := m.mustFindQuestionByName(QuestionName_EnterDomain_Project).Value()
		portal := m.mustFindQuestionByName(QuestionName_EnterDomain_Portal).Value()
		accounts := m.mustFindQuestionByName(QuestionName_EnterDomain_Accounts).Value()

		if accounts == project {
			err := fmt.Errorf("It cannot be equal to %v. Please enter a different value", project)
			m.Questions[len(m.Questions)-1].Model = m.Questions[len(m.Questions)-1].Model.WithError(err)
			return m, true
		} else if accounts == portal {
			err := fmt.Errorf("It cannot be equal to %v. Please enter a different value", portal)
			m.Questions[len(m.Questions)-1].Model = m.Questions[len(m.Questions)-1].Model.WithError(err)
			return m, true
		} else {
			m.Questions[len(m.Questions)-1].Model = m.Questions[len(m.Questions)-1].Model.WithError(nil)
		}
	case QuestionName_EnterAdminPassword_Confirm:
		passwordValue := m.mustFindQuestionByName(QuestionName_EnterAdminPassword).Value()
		confirmValue := m.mustFindQuestionByName(QuestionName_EnterAdminPassword_Confirm).Value()

		if confirmValue != passwordValue {
			err := fmt.Errorf("Passwords mismatch. Please confirm they are the same")
			m.Questions[len(m.Questions)-1].Model = m.Questions[len(m.Questions)-1].Model.WithError(err)
			return m, true
		} else {
			m.Questions[len(m.Questions)-1].Model = m.Questions[len(m.Questions)-1].Model.WithError(nil)
		}
	}

	return m, false
}

func (m SetupApp) blurCurrentQuestion() SetupApp {
	m.Questions[len(m.Questions)-1].Model = m.Questions[len(m.Questions)-1].Model.Blur()
	return m
}

func (m SetupApp) IsCurrentQuestionFocused() bool {
	return m.Questions[len(m.Questions)-1].Model.IsFocused()
}

func (m SetupApp) proceed_QuestionName_EnterDomain_Apex() SetupApp {
	domains := m.ToSuggestedDomains()

	q := newQuestion(Question_ConfirmDefaultDomains)
	questionModel := q.Model.(bubbleteautil.SimplePicker)
	questionModel.Title = fmt.Sprintf(`The following domains will be set up:
- %v
    The authentication endpoint
- %v
    The Authgear portal
- %v
    For logging into the Authgear portal

`, bubbleteautil.StyleForegroundSemanticInfo.Render(domains.Project),
		bubbleteautil.StyleForegroundSemanticInfo.Render(domains.Portal),
		bubbleteautil.StyleForegroundSemanticInfo.Render(domains.Accounts),
	)
	q.Model = questionModel

	m.Questions = append(m.Questions, q)

	return m
}

func (m SetupApp) appendNextQuestionForSetup() (SetupApp, tea.Cmd) {
	if len(m.Questions) == 0 {
		m.Questions = append(m.Questions, newQuestion(Question_AcceptAgreement))
		return m, nil
	}

	// First, we need to perform simple field-level validation on the answer.
	m, earlyReturn := m.performSimpleValidation()
	if earlyReturn {
		return m, nil
	}

	// Second, we need to perform cross-field validation on the particular question.
	m, earlyReturn = m.performCrossFieldValidation()
	if earlyReturn {
		return m, nil
	}

	// When we reach here, the current question is answered.
	// Blur it.
	m = m.blurCurrentQuestion()

	// Finally proceed to next question, if there is any.
	switch m.Questions[len(m.Questions)-1].Name {
	case QuestionName_AcceptAgreement:
		m.Questions = append(m.Questions, newQuestion(Question_EnterDomain_Apex))
	case QuestionName_EnterDomain_Apex:
		m = m.proceed_QuestionName_EnterDomain_Apex()
	case QuestionName_ConfirmDefaultDomains:
		value := m.Questions[len(m.Questions)-1].Value()
		switch value {
		case ValueTrue:
			if m.QuestionName_EnableCertbot_PromptEnabled {
				m.Questions = append(m.Questions, newQuestion(Question_EnableCertbot))
			} else {
				m.Questions = append(m.Questions, newQuestion(Question_EnterAdminEmail))
			}
		case ValueFalse:
			m.Questions = append(m.Questions, newQuestion(Question_EnterDomain_Project))
		}
	case QuestionName_EnterDomain_Project:
		m.Questions = append(m.Questions, newQuestion(Question_EnterDomain_Portal))
	case QuestionName_EnterDomain_Portal:
		m.Questions = append(m.Questions, newQuestion(Question_EnterDomain_Accounts))
	case QuestionName_EnterDomain_Accounts:
		if m.QuestionName_EnableCertbot_PromptEnabled {
			m.Questions = append(m.Questions, newQuestion(Question_EnableCertbot))
		} else {
			m.Questions = append(m.Questions, newQuestion(Question_EnterAdminEmail))
		}
	case QuestionName_EnableCertbot:
		value := m.Questions[len(m.Questions)-1].Value()
		switch value {
		case ValueTrue:
			if m.QuestionName_SelectCertbotEnvironment_PromptEnabled {
				m.Questions = append(m.Questions, newQuestion(Question_SelectCertbotEnvironment))
			} else {
				m.Questions = append(m.Questions, newQuestion(Question_EnterAdminEmail))
			}
		case ValueFalse:
			m.Questions = append(m.Questions, newQuestion(Question_EnterAdminEmail))
		}
	case QuestionName_SelectCertbotEnvironment:
		m.Questions = append(m.Questions, newQuestion(Question_EnterAdminEmail))
	case QuestionName_EnterAdminEmail:
		m.Questions = append(m.Questions, newQuestion(Question_EnterAdminPassword))
	case QuestionName_EnterAdminPassword:
		m.Questions = append(m.Questions, newQuestion(Question_EnterAdminPassword_Confirm))
	case QuestionName_EnterAdminPassword_Confirm:
		m.Questions = append(m.Questions, newQuestion(Question_SelectSMTP))
	case QuestionName_SelectSMTP:
		value := m.Questions[len(m.Questions)-1].Value()
		switch value {
		case SMTPSendgrid:
			m.Questions = append(m.Questions, newQuestion(Question_EnterSendgridAPIKey))
		case SMTPCustom:
			q := newQuestion(Question_EnterSMTPHost)
			q = q.WithValue(m.retainedValues.SMTPHost)
			m.Questions = append(m.Questions, q)
		case SMTPSkip:
			return m, SetupAppEndSurvey
		}
	case QuestionName_EnterSMTPHost:
		q := newQuestion(Question_EnterSMTPPort)
		q = q.WithValue(m.retainedValues.SMTPPort)
		m.Questions = append(m.Questions, q)
	case QuestionName_EnterSMTPPort:
		q := newQuestion(Question_EnterSMTPUsername)
		q = q.WithValue(m.retainedValues.SMTPUsername)
		m.Questions = append(m.Questions, q)
	case QuestionName_EnterSMTPUsername:
		m.Questions = append(m.Questions, newQuestion(Question_EnterSMTPPassword))
	case QuestionName_EnterSMTPPassword, QuestionName_EnterSendgridAPIKey:
		q := newQuestion(Question_EnterSMTPSenderAddress)
		q = q.WithValue(m.retainedValues.SMTPSenderAddress)
		m.Questions = append(m.Questions, q)
	case QuestionName_EnterSMTPSenderAddress:
		q := newQuestion(Question_EnterTestEmailAddress)
		q = q.WithValue(m.retainedValues.TestEmailAddress)
		m.Questions = append(m.Questions, q)
	case QuestionName_EnterTestEmailAddress:
		opts := m.makeSendTestEmailOptions()
		return m, tea.Batch(
			m.StartLoading("Sending email..."),
			func() tea.Msg {
				err := SendTestEmail(opts)
				return msgSendTestEmail{
					Opts: opts,
					Err:  err,
				}
			},
		)
	case QuestionName_AskForTestEmailResult:
		value := m.Questions[len(m.Questions)-1].Value()
		switch value {
		case SendTestEmailResultSuccess:
			return m, SetupAppEndSurvey
		case SendTestEmailResultCorrectSenderAndRetry:
			// Start over from Question_EnterTestEmailAddress
			idx, oldQuestion := m.mustFindQuestionByName_ReturnIndex(QuestionName_EnterTestEmailAddress)
			m.Questions = m.Questions[:idx]
			newQuestion := newQuestion(Question_EnterTestEmailAddress)
			// Retain the value.
			newQuestion = newQuestion.WithValue(oldQuestion.Value())
			m.Questions = append(m.Questions, newQuestion)
		case SendTestEmailResultReconfigureSMTPAndRetry:
			// Start over from Question_SelectSMTP
			m.retainedValues = m.retainValues()
			idx, _ := m.mustFindQuestionByName_ReturnIndex(QuestionName_SelectSMTP)
			m.Questions = m.Questions[:idx]
			newQuestion := newQuestion(Question_SelectSMTP)
			m.Questions = append(m.Questions, newQuestion)
		}
	}

	return m, nil
}

func (m SetupApp) appendNextQuestionForResetup() (SetupApp, tea.Cmd) {
	if len(m.Questions) == 0 {
		q := newQuestion(Question_EnterDomain_Apex)
		questionModel := q.Model.(bubbleteautil.SingleLineTextInput)
		questionModel.Title = "Re-running the installation......\n\n"
		q.Model = questionModel
		m.Questions = append(m.Questions, q)
		return m, nil
	}

	// First, we need to perform simple field-level validation on the answer.
	m, earlyReturn := m.performSimpleValidation()
	if earlyReturn {
		return m, nil
	}

	// Second, we need to perform cross-field validation on the particular question.
	m, earlyReturn = m.performCrossFieldValidation()
	if earlyReturn {
		return m, nil
	}

	// When we reach here, the current question is answered.
	// Blur it.
	m = m.blurCurrentQuestion()

	switch m.Questions[len(m.Questions)-1].Name {
	case QuestionName_EnterDomain_Apex:
		m = m.proceed_QuestionName_EnterDomain_Apex()
	case QuestionName_ConfirmDefaultDomains:
		value := m.Questions[len(m.Questions)-1].Value()
		switch value {
		case ValueTrue:
			if m.QuestionName_EnableCertbot_PromptEnabled {
				m.Questions = append(m.Questions, newQuestion(Question_EnableCertbot))
			} else {
				return m, func() tea.Msg { return msgSetupAppInitResetup{} }
			}
		case ValueFalse:
			m.Questions = append(m.Questions, newQuestion(Question_EnterDomain_Project))
		}
	case QuestionName_EnterDomain_Project:
		m.Questions = append(m.Questions, newQuestion(Question_EnterDomain_Portal))
	case QuestionName_EnterDomain_Portal:
		m.Questions = append(m.Questions, newQuestion(Question_EnterDomain_Accounts))
	case QuestionName_EnterDomain_Accounts:
		if m.QuestionName_EnableCertbot_PromptEnabled {
			m.Questions = append(m.Questions, newQuestion(Question_EnableCertbot))
		} else {
			return m, func() tea.Msg { return msgSetupAppInitResetup{} }
		}
	case QuestionName_EnableCertbot:
		value := m.Questions[len(m.Questions)-1].Value()
		switch value {
		case ValueTrue:
			if m.QuestionName_SelectCertbotEnvironment_PromptEnabled {
				m.Questions = append(m.Questions, newQuestion(Question_SelectCertbotEnvironment))
			} else {
				return m, func() tea.Msg { return msgSetupAppInitResetup{} }
			}
		case ValueFalse:
			return m, func() tea.Msg { return msgSetupAppInitResetup{} }
		}
	case QuestionName_SelectCertbotEnvironment:
		return m, func() tea.Msg { return msgSetupAppInitResetup{} }
	}

	return m, nil
}

func (m SetupApp) appendNextQuestion() (SetupApp, tea.Cmd) {
	if m.IsResetup {
		return m.appendNextQuestionForResetup()
	} else {
		return m.appendNextQuestionForSetup()
	}
}

func (m SetupApp) findQuestionByName(name QuestionName) (*int, *Question, bool) {
	for idx, q := range m.Questions {
		if q.Name == name {
			idx := idx
			q := q
			return &idx, &q, true
		}
	}
	return nil, nil, false
}

func (m SetupApp) mustFindQuestionByName_ReturnIndex(name QuestionName) (int, Question) {
	idx, q, ok := m.findQuestionByName(name)
	if !ok {
		panic(fmt.Errorf("question not found: %v", name))
	}
	return *idx, *q
}

func (m SetupApp) mustFindQuestionByName(name QuestionName) Question {
	_, q := m.mustFindQuestionByName_ReturnIndex(name)
	return q
}

func (m SetupApp) retainValues() RetainedValues {
	values := RetainedValues{}
	if _, smtpHost, ok := m.findQuestionByName(QuestionName_EnterSMTPHost); ok {
		values.SMTPHost = smtpHost.Value()
	}
	if _, smtpPort, ok := m.findQuestionByName(QuestionName_EnterSMTPPort); ok {
		values.SMTPPort = smtpPort.Value()
	}
	if _, smtpUsername, ok := m.findQuestionByName(QuestionName_EnterSMTPUsername); ok {
		values.SMTPUsername = smtpUsername.Value()
	}
	if _, smtpSenderAddress, ok := m.findQuestionByName(QuestionName_EnterSMTPSenderAddress); ok {
		values.SMTPSenderAddress = smtpSenderAddress.Value()
	}
	if _, testEmailAddress, ok := m.findQuestionByName(QuestionName_EnterTestEmailAddress); ok {
		values.TestEmailAddress = testEmailAddress.Value()
	}
	return values
}

func (m SetupApp) makeSendTestEmailOptions() SendTestEmailOptions {
	opts := SendTestEmailOptions{
		SenderAddress: m.mustFindQuestionByName(QuestionName_EnterSMTPSenderAddress).Value(),
		ToAddress:     m.mustFindQuestionByName(QuestionName_EnterTestEmailAddress).Value(),
	}
	switch m.mustFindQuestionByName(QuestionName_SelectSMTP).Value() {
	case SMTPSendgrid:
		opts.Host = "smtp.sendgrid.net"
		opts.Port = 587
		opts.Username = "apikey"
		opts.Password = m.mustFindQuestionByName(QuestionName_EnterSendgridAPIKey).Value()
	case SMTPCustom:
		opts.Host = m.mustFindQuestionByName(QuestionName_EnterSMTPHost).Value()
		opts.Port, _ = strconv.Atoi(m.mustFindQuestionByName(QuestionName_EnterSMTPPort).Value())
		opts.Username = m.mustFindQuestionByName(QuestionName_EnterSMTPUsername).Value()
		opts.Password = m.mustFindQuestionByName(QuestionName_EnterSMTPPassword).Value()
	}
	return opts
}

func (m *SetupApp) StartLoading(msg string) tea.Cmd {
	m.Loading = true
	m.LoadingMessage = msg
	m.Spinner = spinner.New()
	m.Spinner.Spinner = spinner.Dot
	m.Spinner.Style = bubbleteautil.StyleForegroundSemanticInfo
	return m.Spinner.Tick
}

func (m *SetupApp) StopLoading() {
	m.Loading = false
}

func (m SetupApp) View() string {
	var b strings.Builder
	for _, q := range m.Questions {
		fmt.Fprintf(&b, "%v\n", q.View())
	}
	if m.Loading {
		fmt.Fprintf(&b, "%v %v\n", m.Spinner.View(), m.LoadingMessage)
	}
	if m.RecoverableErr != nil {
		fmt.Fprintf(&b, "❌ Encountered this error\n%v\n\n%v\n",
			bubbleteautil.StyleForegroundSemanticError.Render(m.RecoverableErr.Error()),
			bubbleteautil.StyleForegroundSemanticInfo.Render("Please hit enter to continue"),
		)
	}

	// FatalError outputs correct newlines.
	fmt.Fprintf(&b, "%v", m.FatalError.View())
	// Installation outputs correct newlines.
	if m.Installation != nil {
		fmt.Fprintf(&b, "%v", m.Installation.View())
	}
	// Resetup outputs correct newlines.
	if m.Resetup != nil {
		fmt.Fprintf(&b, "%v", m.Resetup.View())
	}
	return b.String()
}

func (m SetupApp) ToSuggestedDomains() Domains {
	apexDomain := m.mustFindQuestionByName(QuestionName_EnterDomain_Apex).Value()
	return Domains{
		Project:  fmt.Sprintf("auth.%v", apexDomain),
		Portal:   fmt.Sprintf("authgear-portal.%v", apexDomain),
		Accounts: fmt.Sprintf("authgear-portal-accounts.%v", apexDomain),
	}
}

func (m SetupApp) ToDomains() Domains {
	confirmSuggestedDomains := m.mustFindQuestionByName(QuestionName_ConfirmDefaultDomains).Value() == ValueTrue
	if confirmSuggestedDomains {
		return m.ToSuggestedDomains()
	}

	return Domains{
		Project:  m.mustFindQuestionByName(QuestionName_EnterDomain_Project).Value(),
		Portal:   m.mustFindQuestionByName(QuestionName_EnterDomain_Portal).Value(),
		Accounts: m.mustFindQuestionByName(QuestionName_EnterDomain_Accounts).Value(),
	}
}

func (m SetupApp) ToInstallation(licenseObject *internal.LicenseObject) Installation {
	certbotEnabled := m.AUTHGEAR_CERTBOT_ENABLED
	if _, q, ok := m.findQuestionByName(QuestionName_EnableCertbot); ok {
		certbotEnabled = q.Value() == ValueTrue
	}

	certbotEnvironment := m.AUTHGEAR_CERTBOT_ENVIRONMENT
	if _, q, ok := m.findQuestionByName(QuestionName_SelectCertbotEnvironment); ok {
		certbotEnvironment = q.Value()
	}

	installation := Installation{
		Context:                           m.Context,
		AUTHGEAR_ONCE_IMAGE:               m.AUTHGEAR_ONCE_IMAGE,
		AUTHGEAR_ONCE_MACHINE_FINGERPRINT: m.AUTHGEAR_ONCE_MACHINE_FINGERPRINT,
		AUTHGEAR_ONCE_LICENSE_KEY:         m.AUTHGEAR_ONCE_LICENSE_KEY,
		AUTHGEAR_ONCE_LICENSE_EXPIRE_AT:   nilTimeToEmptyString(licenseObject.ExpireAt),
		AUTHGEAR_ONCE_LICENSEE_EMAIL:      nilStringToEmptyString(licenseObject.LicenseeEmail),

		AUTHGEAR_ONCE_ADMIN_USER_EMAIL:    m.mustFindQuestionByName(QuestionName_EnterAdminEmail).Value(),
		AUTHGEAR_ONCE_ADMIN_USER_PASSWORD: m.mustFindQuestionByName(QuestionName_EnterAdminPassword).Value(),

		AUTHGEAR_CERTBOT_ENABLED:     strconv.FormatBool(certbotEnabled),
		AUTHGEAR_CERTBOT_ENVIRONMENT: certbotEnvironment,
	}

	domains := m.ToDomains()

	installation.AUTHGEAR_HTTP_ORIGIN_PROJECT = fmt.Sprintf("%v://%v", m.HTTPScheme, domains.Project)
	installation.AUTHGEAR_HTTP_ORIGIN_PORTAL = fmt.Sprintf("%v://%v", m.HTTPScheme, domains.Portal)
	installation.AUTHGEAR_HTTP_ORIGIN_ACCOUNTS = fmt.Sprintf("%v://%v", m.HTTPScheme, domains.Accounts)

	switch m.mustFindQuestionByName(QuestionName_SelectSMTP).Value() {
	case SMTPSkip:
		break
	default:
		opts := m.makeSendTestEmailOptions()
		installation.AUTHGEAR_SMTP_HOST = opts.Host
		installation.AUTHGEAR_SMTP_PORT = opts.Port
		installation.AUTHGEAR_SMTP_USERNAME = opts.Username
		installation.AUTHGEAR_SMTP_PASSWORD = opts.Password
		installation.AUTHGEAR_SMTP_SENDER_ADDRESS = opts.SenderAddress
	}
	return installation
}

func (m SetupApp) ToResetup(licenseOptions internal.LicenseOptions) Resetup {
	certbotEnabled := m.AUTHGEAR_CERTBOT_ENABLED
	if _, q, ok := m.findQuestionByName(QuestionName_EnableCertbot); ok {
		certbotEnabled = q.Value() == ValueTrue
	}

	certbotEnvironment := m.AUTHGEAR_CERTBOT_ENVIRONMENT
	if _, q, ok := m.findQuestionByName(QuestionName_SelectCertbotEnvironment); ok {
		certbotEnvironment = q.Value()
	}

	resetup := Resetup{
		Context:                      m.Context,
		AUTHGEAR_ONCE_LICENSE_KEY:    licenseOptions.LicenseKey,
		AUTHGEAR_ONCE_IMAGE:          m.AUTHGEAR_ONCE_IMAGE,
		AUTHGEAR_CERTBOT_ENABLED:     strconv.FormatBool(certbotEnabled),
		AUTHGEAR_CERTBOT_ENVIRONMENT: certbotEnvironment,
	}

	domains := m.ToDomains()

	resetup.AUTHGEAR_HTTP_ORIGIN_PROJECT = fmt.Sprintf("%v://%v", m.HTTPScheme, domains.Project)
	resetup.AUTHGEAR_HTTP_ORIGIN_PORTAL = fmt.Sprintf("%v://%v", m.HTTPScheme, domains.Portal)
	resetup.AUTHGEAR_HTTP_ORIGIN_ACCOUNTS = fmt.Sprintf("%v://%v", m.HTTPScheme, domains.Accounts)

	return resetup
}

func (m SetupApp) HasError() bool {
	if m.FatalError.Err != nil {
		return true
	}
	if m.Installation != nil {
		if m.Installation.FatalError.Err != nil {
			return true
		}
	}
	return false
}

func newQuestion(q Question) Question {
	// q is copied.
	q.Model = q.Model.Focus()
	return q
}

type msgInstallationInstall struct {
	Err error
}

type msgInstallationStart struct {
	Err error
}

type InstallationStatus int

const (
	InstallationStatusInstalling InstallationStatus = iota
	InstallationStatusStarting
	InstallationStatusDone
)

type Domains struct {
	Project  string
	Portal   string
	Accounts string
}

type Installation struct {
	Context context.Context

	AUTHGEAR_ONCE_IMAGE               string
	AUTHGEAR_ONCE_MACHINE_FINGERPRINT string
	AUTHGEAR_ONCE_LICENSE_KEY         string
	AUTHGEAR_ONCE_LICENSE_EXPIRE_AT   string
	AUTHGEAR_ONCE_LICENSEE_EMAIL      string

	AUTHGEAR_HTTP_ORIGIN_PROJECT      string
	AUTHGEAR_HTTP_ORIGIN_PORTAL       string
	AUTHGEAR_HTTP_ORIGIN_ACCOUNTS     string
	AUTHGEAR_ONCE_ADMIN_USER_EMAIL    string
	AUTHGEAR_ONCE_ADMIN_USER_PASSWORD string
	AUTHGEAR_CERTBOT_ENABLED          string
	AUTHGEAR_CERTBOT_ENVIRONMENT      string
	AUTHGEAR_SMTP_HOST                string
	AUTHGEAR_SMTP_PORT                int
	AUTHGEAR_SMTP_USERNAME            string
	AUTHGEAR_SMTP_PASSWORD            string
	AUTHGEAR_SMTP_SENDER_ADDRESS      string

	Spinner            spinner.Model
	InstallationStatus InstallationStatus
	Loading            bool

	FatalError internal.FatalError
}

var _ tea.Model = Installation{}

func (m Installation) Init() tea.Cmd {
	return nil
}

func (m Installation) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case msgSetupStartInstallation:
		m.Spinner = spinner.New()
		m.Spinner.Spinner = spinner.Dot
		m.Spinner.Style = bubbleteautil.StyleForegroundSemanticInfo

		dockerRunOptions := newDockerRunOptionsForInstallation(m)
		m.Loading = true
		cmds = append(cmds, m.Spinner.Tick, func() tea.Msg {
			// `docker run` is smart enough to pull the image and create the volume if the volume does not exist.
			// So we do not need to do that manually.
			// In fact, if we `docker pull`, it will result in error if we pull a image that exists only locally.
			_, err := internal.DockerRunWithCertbotErrorHandling(m.Context, dockerRunOptions)
			if errors.Is(err, internal.ErrCertbotExitCode10) {
				err = errors.Join(&internal.ErrCertbotFailedToGetCertificates{
					Domains: m.ToDomains(),
				}, err)
			}
			return msgInstallationInstall{
				Err: err,
			}
		})
	case msgInstallationInstall:
		m.Loading = false
		if msg.Err != nil {
			m.FatalError = m.FatalError.WithErr(msg.Err)
			return m, tea.Quit
		} else {
			m.Loading = true
			m.InstallationStatus = InstallationStatusStarting
			dockerRunOptions := internal.NewDockerRunOptionsForStarting(m.AUTHGEAR_ONCE_IMAGE)
			cmds = append(cmds, func() tea.Msg {
				_, err := internal.DockerRun(m.Context, dockerRunOptions)
				return msgInstallationStart{
					Err: err,
				}
			})
		}
	case msgInstallationStart:
		m.Loading = false
		if msg.Err != nil {
			m.FatalError = m.FatalError.WithErr(msg.Err)
			return m, tea.Quit
		} else {
			m.InstallationStatus = InstallationStatusDone
			return m, tea.Quit
		}
	}

	var updated spinner.Model
	var newCmd tea.Cmd
	updated, newCmd = m.Spinner.Update(msg)
	m.Spinner = updated
	cmds = append(cmds, newCmd)

	return m, tea.Batch(cmds...)
}

func (m Installation) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "The installation is going to take few minutes:\n")
	switch m.InstallationStatus {
	case InstallationStatusInstalling:
		var spinner string
		if m.Loading {
			spinner = fmt.Sprintf(" %v", m.Spinner.View())
		}
		fmt.Fprintf(&b, "  Installing%v\n", spinner)
	case InstallationStatusStarting:
		fmt.Fprintf(&b, "  Installed\n")
		var spinner string
		if m.Loading {
			spinner = fmt.Sprintf(" %v", m.Spinner.View())
		}
		fmt.Fprintf(&b, "  Starting%v\n", spinner)
	case InstallationStatusDone:
		fmt.Fprintf(&b, "  Installed\n")
		fmt.Fprintf(&b, "  Started\n")
		fmt.Fprintf(&b, "\n")
		if m.AUTHGEAR_CERTBOT_ENABLED == "true" {
			fmt.Fprintf(&b, "Generated TLS certificates issued by Let's Encrypt.\n")
		}
		fmt.Fprintf(
			&b,
			"Ready! Start using Authear by visiting\n\n  %v\n\n",
			bubbleteautil.StyleForegroundSemanticInfo.Render(m.AUTHGEAR_HTTP_ORIGIN_PORTAL),
		)
	}

	fmt.Fprintf(&b, "%v", m.FatalError.View())

	return b.String()
}

func (m Installation) ToDomains() []string {
	return originsToDomains(
		m.AUTHGEAR_HTTP_ORIGIN_PROJECT,
		m.AUTHGEAR_HTTP_ORIGIN_PORTAL,
		m.AUTHGEAR_HTTP_ORIGIN_ACCOUNTS,
	)
}

func newDockerRunOptionsForInstallation(m Installation) internal.DockerRunOptions {
	opts := internal.NewDockerRunOptionsForStarting(m.AUTHGEAR_ONCE_IMAGE)
	opts.Detach = false
	// No need to specify --restart as this is a short-lived container.
	opts.Restart = ""
	// Block on getting the TLS certificates.
	opts.Command = []string{"docker_wrapper", "--block-on-getting-tls-certificates"}
	// Keep or remove the container by a flag.
	opts.Rm = !internal.KeepInstallationContainerByDefault
	// Do not specify name so that it will not clash with the actual container.
	opts.Name = ""
	opts.Env = []string{
		fmt.Sprintf("AUTHGEAR_ONCE_IMAGE=%v", m.AUTHGEAR_ONCE_IMAGE),
		fmt.Sprintf("AUTHGEAR_ONCE_LICENSE_KEY=%v", m.AUTHGEAR_ONCE_LICENSE_KEY),
		fmt.Sprintf("AUTHGEAR_ONCE_LICENSE_EXPIRE_AT=%v", m.AUTHGEAR_ONCE_LICENSE_EXPIRE_AT),
		fmt.Sprintf("AUTHGEAR_ONCE_LICENSEE_EMAIL=%v", m.AUTHGEAR_ONCE_LICENSEE_EMAIL),
		fmt.Sprintf("AUTHGEAR_ONCE_MACHINE_FINGERPRINT=%v", m.AUTHGEAR_ONCE_MACHINE_FINGERPRINT),
		fmt.Sprintf("AUTHGEAR_HTTP_ORIGIN_PROJECT=%v", m.AUTHGEAR_HTTP_ORIGIN_PROJECT),
		fmt.Sprintf("AUTHGEAR_HTTP_ORIGIN_PORTAL=%v", m.AUTHGEAR_HTTP_ORIGIN_PORTAL),
		fmt.Sprintf("AUTHGEAR_HTTP_ORIGIN_ACCOUNTS=%v", m.AUTHGEAR_HTTP_ORIGIN_ACCOUNTS),
		fmt.Sprintf("AUTHGEAR_ONCE_ADMIN_USER_EMAIL=%v", m.AUTHGEAR_ONCE_ADMIN_USER_EMAIL),
		fmt.Sprintf("AUTHGEAR_ONCE_ADMIN_USER_PASSWORD=%v", m.AUTHGEAR_ONCE_ADMIN_USER_PASSWORD),
		fmt.Sprintf("AUTHGEAR_CERTBOT_ENABLED=%v", m.AUTHGEAR_CERTBOT_ENABLED),
		fmt.Sprintf("AUTHGEAR_CERTBOT_ENVIRONMENT=%v", m.AUTHGEAR_CERTBOT_ENVIRONMENT),
	}
	if m.AUTHGEAR_SMTP_HOST != "" {
		opts.Env = append(opts.Env,
			fmt.Sprintf("AUTHGEAR_SMTP_HOST=%v", m.AUTHGEAR_SMTP_HOST),
			fmt.Sprintf("AUTHGEAR_SMTP_PORT=%v", m.AUTHGEAR_SMTP_PORT),
			fmt.Sprintf("AUTHGEAR_SMTP_USERNAME=%v", m.AUTHGEAR_SMTP_USERNAME),
			fmt.Sprintf("AUTHGEAR_SMTP_PASSWORD=%v", m.AUTHGEAR_SMTP_PASSWORD),
			fmt.Sprintf("AUTHGEAR_SMTP_SENDER_ADDRESS=%v", m.AUTHGEAR_SMTP_SENDER_ADDRESS),
		)
	}
	return opts
}

func newDockerRunOptionsForResetup(m Resetup) internal.DockerRunOptions {
	opts := internal.NewDockerRunOptionsForStarting(m.AUTHGEAR_ONCE_IMAGE)
	opts.Detach = false
	// No need to specify --restart as this is a short-lived container.
	opts.Restart = ""
	// Block on getting the TLS certificates.
	opts.Command = []string{"docker_wrapper", "--block-on-getting-tls-certificates"}
	// Keep or remove the container by a flag.
	opts.Rm = !internal.KeepInstallationContainerByDefault
	// Do not specify name so that it will not clash with the actual container.
	opts.Name = ""
	opts.Env = []string{
		fmt.Sprintf("AUTHGEAR_ONCE_IMAGE=%v", m.AUTHGEAR_ONCE_IMAGE),
		fmt.Sprintf("AUTHGEAR_HTTP_ORIGIN_PROJECT=%v", m.AUTHGEAR_HTTP_ORIGIN_PROJECT),
		fmt.Sprintf("AUTHGEAR_HTTP_ORIGIN_PORTAL=%v", m.AUTHGEAR_HTTP_ORIGIN_PORTAL),
		fmt.Sprintf("AUTHGEAR_HTTP_ORIGIN_ACCOUNTS=%v", m.AUTHGEAR_HTTP_ORIGIN_ACCOUNTS),
		fmt.Sprintf("AUTHGEAR_CERTBOT_ENABLED=%v", m.AUTHGEAR_CERTBOT_ENABLED),
		fmt.Sprintf("AUTHGEAR_CERTBOT_ENVIRONMENT=%v", m.AUTHGEAR_CERTBOT_ENVIRONMENT),
	}
	return opts
}

type msgResetupResult struct {
	Err error
}

type msgResetupStart struct {
	Err error
}

type Resetup struct {
	Context context.Context

	AUTHGEAR_ONCE_LICENSE_KEY     string
	AUTHGEAR_ONCE_IMAGE           string
	AUTHGEAR_HTTP_ORIGIN_PROJECT  string
	AUTHGEAR_HTTP_ORIGIN_PORTAL   string
	AUTHGEAR_HTTP_ORIGIN_ACCOUNTS string

	AUTHGEAR_CERTBOT_ENABLED     string
	AUTHGEAR_CERTBOT_ENVIRONMENT string

	Spinner            spinner.Model
	InstallationStatus InstallationStatus
	Loading            bool

	FatalError internal.FatalError
}

var _ tea.Model = Resetup{}

func (m Resetup) Init() tea.Cmd {
	return nil
}

func (m Resetup) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case msgSetupAppStartResetup:
		m.Spinner = spinner.New()
		m.Spinner.Spinner = spinner.Dot
		m.Spinner.Style = bubbleteautil.StyleForegroundSemanticInfo

		dockerRunOptions := newDockerRunOptionsForResetup(m)
		m.Loading = true
		cmds = append(cmds, m.Spinner.Tick, func() tea.Msg {
			// docker rm -f authgearonce
			err := internal.DockerRm(m.Context, internal.NameDockerContainer, internal.DockerRmOptions{
				Force: true,
			})
			if err != nil {
				return msgResetupResult{
					Err: err,
				}
			}

			// docker run
			_, err = internal.DockerRunWithCertbotErrorHandling(m.Context, dockerRunOptions)
			if errors.Is(err, internal.ErrCertbotExitCode10) {
				err = errors.Join(&internal.ErrCertbotFailedToGetCertificates{
					Domains: m.ToDomains(),
				}, err)
			}
			if err != nil {
				return msgResetupResult{
					Err: err,
				}
			}

			return msgResetupResult{}
		})
	case msgResetupResult:
		m.Loading = false
		if msg.Err != nil {
			m.FatalError = m.FatalError.WithErr(msg.Err)
			return m, tea.Quit
		} else {
			m.Loading = true
			m.InstallationStatus = InstallationStatusStarting
			dockerRunOptions := internal.NewDockerRunOptionsForStarting(m.AUTHGEAR_ONCE_IMAGE)
			cmds = append(cmds, func() tea.Msg {
				_, err := internal.DockerRun(m.Context, dockerRunOptions)
				return msgResetupStart{
					Err: err,
				}
			})
		}
	case msgResetupStart:
		m.Loading = false
		if msg.Err != nil {
			m.FatalError = m.FatalError.WithErr(msg.Err)
			return m, tea.Quit
		} else {
			m.InstallationStatus = InstallationStatusDone
			return m, tea.Quit
		}
	}

	var updated spinner.Model
	var newCmd tea.Cmd
	updated, newCmd = m.Spinner.Update(msg)
	m.Spinner = updated
	cmds = append(cmds, newCmd)

	return m, tea.Batch(cmds...)
}

func (m Resetup) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "The installation is going to take few minutes:\n")
	switch m.InstallationStatus {
	case InstallationStatusInstalling:
		var spinner string
		if m.Loading {
			spinner = fmt.Sprintf(" %v", m.Spinner.View())
		}
		fmt.Fprintf(&b, "  Installing%v\n", spinner)
	case InstallationStatusStarting:
		fmt.Fprintf(&b, "  Installed\n")
		var spinner string
		if m.Loading {
			spinner = fmt.Sprintf(" %v", m.Spinner.View())
		}
		fmt.Fprintf(&b, "  Starting%v\n", spinner)
	case InstallationStatusDone:
		fmt.Fprintf(&b, "  Installed\n")
		fmt.Fprintf(&b, "  Started\n")
		fmt.Fprintf(&b, "\n")
		if m.AUTHGEAR_CERTBOT_ENABLED == "true" {
			fmt.Fprintf(&b, "Generated TLS certificates issued by Let's Encrypt.\n")
		}
		fmt.Fprintf(
			&b,
			"Ready! Start using Authear by visiting\n\n  %v\n\n",
			bubbleteautil.StyleForegroundSemanticInfo.Render(m.AUTHGEAR_HTTP_ORIGIN_PORTAL),
		)
	}

	fmt.Fprintf(&b, "%v", m.FatalError.View())

	return b.String()
}

func (m Resetup) ToDomains() []string {
	return originsToDomains(
		m.AUTHGEAR_HTTP_ORIGIN_PROJECT,
		m.AUTHGEAR_HTTP_ORIGIN_PORTAL,
		m.AUTHGEAR_HTTP_ORIGIN_ACCOUNTS,
	)
}

func nilStringToEmptyString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func nilTimeToEmptyString(t *time.Time) string {
	if t == nil {
		return ""
	}

	return t.Format(time.RFC3339)
}

func originsToDomains(origins ...string) []string {
	var domains []string
	for _, origin := range origins {
		u, err := url.Parse(origin)
		if err != nil {
			panic(err)
		}
		domains = append(domains, u.Host)
	}
	return domains
}
