package cmdsetup

import (
	"context"
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"strings"

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
	Context context.Context
	Image   string

	Questions      []Question
	retainedValues RetainedValues

	Loading        bool
	LoadingMessage string
	Spinner        spinner.Model

	FatalError internal.FatalError

	RecoverableErr    error
	RecoverableErrCmd tea.Cmd

	Installation *Installation
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

type msgSetupStartInstallation struct{}

func SetupAppStartInstallation() tea.Msg {
	return msgSetupStartInstallation{}
}

func (m SetupApp) Init() tea.Cmd {
	return SetupAppInit
}

func (m SetupApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case msgSetupAppAbort:
		return m, tea.Quit
	case msgSetupAppInit:
		_, err := exec.LookPath(internal.BinDocker)
		if err != nil {
			m.FatalError = m.FatalError.WithErr(internal.ErrNoDocker)
			return m, tea.Quit
		}

		volumes, err := internal.DockerVolumeLs(m.Context)
		if err != nil {
			m.FatalError = m.FatalError.WithErr(err)
			return m, tea.Quit
		}

		if slices.ContainsFunc(volumes, func(v internal.DockerVolume) bool {
			return v.Name == internal.NameDockerVolume && v.Scope == internal.DockerVolumeScopeLocal
		}) {
			m.FatalError = m.FatalError.WithErr(internal.ErrDockerVolumeExists)
			return m, tea.Quit
		}

		return m, SetupAppStartSurvey
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
			} else {
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
		installation := m.ToInstallation()
		m.Installation = &installation
		return m, SetupAppStartInstallation
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

	return m, tea.Batch(cmds...)
}

func (m SetupApp) appendNextQuestion() (SetupApp, tea.Cmd) {
	if len(m.Questions) == 0 {
		m.Questions = append(m.Questions, newQuestion(Question_AcceptAgreement))
		return m, nil
	}

	// First, we need to perform simple field-level validation on the answer.
	updated, valid := m.Questions[len(m.Questions)-1].Model.Validate()
	m.Questions[len(m.Questions)-1].Model = updated.(bubbleteautil.Model)
	if !valid {
		return m, nil
	}

	// Second, we need to perform cross-field validation on the particular question.
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
			return m, nil
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
			return m, nil
		} else if accounts == portal {
			err := fmt.Errorf("It cannot be equal to %v. Please enter a different value", portal)
			m.Questions[len(m.Questions)-1].Model = m.Questions[len(m.Questions)-1].Model.WithError(err)
			return m, nil
		} else {
			m.Questions[len(m.Questions)-1].Model = m.Questions[len(m.Questions)-1].Model.WithError(nil)
		}
	case QuestionName_EnterAdminPassword_Confirm:
		passwordValue := m.mustFindQuestionByName(QuestionName_EnterAdminPassword).Value()
		confirmValue := m.mustFindQuestionByName(QuestionName_EnterAdminPassword_Confirm).Value()

		if confirmValue != passwordValue {
			err := fmt.Errorf("Passwords mismatch. Please confirm they are the same")
			m.Questions[len(m.Questions)-1].Model = m.Questions[len(m.Questions)-1].Model.WithError(err)
			return m, nil
		} else {
			m.Questions[len(m.Questions)-1].Model = m.Questions[len(m.Questions)-1].Model.WithError(nil)
		}
	}

	// When we reach here, the current question is answered.
	// Blur it.
	m.Questions[len(m.Questions)-1].Model = m.Questions[len(m.Questions)-1].Model.Blur()

	// Finally proceed to next question, if there is any.
	switch m.Questions[len(m.Questions)-1].Name {
	case QuestionName_AcceptAgreement:
		m.Questions = append(m.Questions, newQuestion(Question_EnterDomain_Apex))
	case QuestionName_EnterDomain_Apex:
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
	case QuestionName_ConfirmDefaultDomains:
		value := m.Questions[len(m.Questions)-1].Value()
		switch value {
		case ValueTrue:
			m.Questions = append(m.Questions, newQuestion(Question_EnableCertbot))
		case ValueFalse:
			m.Questions = append(m.Questions, newQuestion(Question_EnterDomain_Project))
		}
	case QuestionName_EnterDomain_Project:
		m.Questions = append(m.Questions, newQuestion(Question_EnterDomain_Portal))
	case QuestionName_EnterDomain_Portal:
		m.Questions = append(m.Questions, newQuestion(Question_EnterDomain_Accounts))
	case QuestionName_EnterDomain_Accounts:
		m.Questions = append(m.Questions, newQuestion(Question_EnableCertbot))
	case QuestionName_EnableCertbot:
		value := m.Questions[len(m.Questions)-1].Value()
		switch value {
		case ValueTrue:
			m.Questions = append(m.Questions, newQuestion(Question_SelectCertbotEnvironment))
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
		fmt.Fprintf(&b, "❌ Encountered this error\n%v\n%v\n",
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

func (m SetupApp) ToInstallation() Installation {
	certbotEnabled := m.mustFindQuestionByName(QuestionName_EnableCertbot).Value() == ValueTrue

	installation := Installation{
		Context:                           m.Context,
		Image:                             m.Image,
		AUTHGEAR_ONCE_ADMIN_USER_EMAIL:    m.mustFindQuestionByName(QuestionName_EnterAdminEmail).Value(),
		AUTHGEAR_ONCE_ADMIN_USER_PASSWORD: m.mustFindQuestionByName(QuestionName_EnterAdminPassword).Value(),
	}

	domains := m.ToDomains()

	scheme := "http"
	if certbotEnabled {
		scheme = "https"
		installation.AUTHGEAR_CERTBOT_ENVIRONMENT = m.mustFindQuestionByName(QuestionName_SelectCertbotEnvironment).Value()
	}
	installation.AUTHGEAR_HTTP_ORIGIN_PROJECT = fmt.Sprintf("%v://%v", scheme, domains.Project)
	installation.AUTHGEAR_HTTP_ORIGIN_PORTAL = fmt.Sprintf("%v://%v", scheme, domains.Portal)
	installation.AUTHGEAR_HTTP_ORIGIN_ACCOUNTS = fmt.Sprintf("%v://%v", scheme, domains.Accounts)

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
	Image   string

	AUTHGEAR_HTTP_ORIGIN_PROJECT      string
	AUTHGEAR_HTTP_ORIGIN_PORTAL       string
	AUTHGEAR_HTTP_ORIGIN_ACCOUNTS     string
	AUTHGEAR_ONCE_ADMIN_USER_EMAIL    string
	AUTHGEAR_ONCE_ADMIN_USER_PASSWORD string
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
			err := internal.DockerRun(m.Context, dockerRunOptions)
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
			dockerRunOptions := internal.NewDockerRunOptionsForStarting(m.Image)
			cmds = append(cmds, func() tea.Msg {
				err := internal.DockerRun(m.Context, dockerRunOptions)
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
		fmt.Fprintf(
			&b,
			"\nReady! Start using Authear by visiting\n\n  %v\n\n",
			bubbleteautil.StyleForegroundSemanticInfo.Render(m.AUTHGEAR_HTTP_ORIGIN_PORTAL),
		)
	}

	fmt.Fprintf(&b, "%v", m.FatalError.View())

	return b.String()
}

func newDockerRunOptionsForInstallation(m Installation) internal.DockerRunOptions {
	opts := internal.NewDockerRunOptionsForStarting(m.Image)
	opts.Detach = false
	// Run the shell command true to exit 0 when container has finished first run.
	opts.Command = []string{"true"}
	// Remove the container because this container always run `true`.
	opts.Rm = true
	opts.Env = []string{
		fmt.Sprintf("AUTHGEAR_HTTP_ORIGIN_PROJECT=%v", m.AUTHGEAR_HTTP_ORIGIN_PROJECT),
		fmt.Sprintf("AUTHGEAR_HTTP_ORIGIN_PORTAL=%v", m.AUTHGEAR_HTTP_ORIGIN_PORTAL),
		fmt.Sprintf("AUTHGEAR_HTTP_ORIGIN_ACCOUNTS=%v", m.AUTHGEAR_HTTP_ORIGIN_ACCOUNTS),
		fmt.Sprintf("AUTHGEAR_ONCE_ADMIN_USER_EMAIL=%v", m.AUTHGEAR_ONCE_ADMIN_USER_EMAIL),
		fmt.Sprintf("AUTHGEAR_ONCE_ADMIN_USER_PASSWORD=%v", m.AUTHGEAR_ONCE_ADMIN_USER_PASSWORD),
	}
	if m.AUTHGEAR_CERTBOT_ENVIRONMENT != "" {
		opts.Env = append(opts.Env, fmt.Sprintf("AUTHGEAR_CERTBOT_ENVIRONMENT=%v", m.AUTHGEAR_CERTBOT_ENVIRONMENT))
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
