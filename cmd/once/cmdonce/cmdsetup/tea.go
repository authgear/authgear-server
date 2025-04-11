package cmdsetup

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

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
	Complete bool

	Questions      []Question
	retainedValues RetainedValues

	Loading        bool
	LoadingMessage string
	Spinner        spinner.Model

	Err           error
	ErrRecoverCmd tea.Cmd
}

var _ tea.Model = SetupApp{}

type msgSetupAppInit struct{}

func SetupAppInit() tea.Msg {
	return msgSetupAppInit{}
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

type msgSetupAppFinish struct{}

func SetupAppFinish() tea.Msg {
	return msgSetupAppFinish{}
}

func (m SetupApp) Init() tea.Cmd {
	return SetupAppInit
}

func (m SetupApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case msgSetupAppInit:
		var cmd tea.Cmd
		m, cmd = m.appendNextQuestion()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case msgSetupAppAbort:
		return m, tea.Quit
	case msgSetupAppFinish:
		m.Complete = true
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, SetupAppAbort
		case tea.KeyEnter:
			if m.Err != nil {
				m.Err = nil
				if m.ErrRecoverCmd != nil {
					cmds = append(cmds, m.ErrRecoverCmd)
					m.ErrRecoverCmd = nil
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
			m.Err = msg.Err
			m.ErrRecoverCmd = errRecoverCmd
		} else {
			cmds = append(cmds, errRecoverCmd)
		}
	case msgAskForSendTestEmailResult:
		q := newQuestion(Question_AskForTestEmailResult)
		picker := q.Model.(bubbleteautil.SimplePicker)
		picker.Prompt = fmt.Sprintf("A test email is sent to %v, did you receive it", msg.Opts.ToAddress)
		q.Model = picker
		m.Questions = append(m.Questions, q)
	}

	for idx := range m.Questions {
		var updated tea.Model
		var newCmd tea.Cmd
		updated, newCmd = m.Questions[idx].Update(msg)
		m.Questions[idx] = updated.(Question)
		cmds = append(cmds, newCmd)
	}
	if m.Loading {
		var updated spinner.Model
		var newCmd tea.Cmd
		updated, newCmd = m.Spinner.Update(msg)
		m.Spinner = updated
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
	case QuestionName_EnterAdminPassword_Confirm:
		passwordValue := m.mustFindQuestionByName(QuestionName_EnterAdminPassword).Value()

		confirmModel := m.Questions[len(m.Questions)-1].Model.(bubbleteautil.SingleLineTextInput)
		confirmValue := confirmModel.Value()

		var err error
		if confirmValue != passwordValue {
			err = fmt.Errorf("Passwords mismatch. Please confirm they are the same")
		}

		confirmModel = confirmModel.WithError(err).(bubbleteautil.SingleLineTextInput)
		m.Questions[len(m.Questions)-1].Model = confirmModel
		if err != nil {
			return m, nil
		}
	}

	// When we reach here, the current question is answered.
	// Blur it.
	m.Questions[len(m.Questions)-1].Model = m.Questions[len(m.Questions)-1].Model.Blur()

	// Finally proceed to next question, if there is any.
	switch m.Questions[len(m.Questions)-1].Name {
	case QuestionName_AcceptAgreement:
		m.Questions = append(m.Questions, newQuestion(Question_EnterDomain_Project))
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
			return m, SetupAppFinish
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
			return m, SetupAppFinish
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
	if m.Err != nil {
		fmt.Fprintf(&b, "❌ Encountered this error\n%v\n%v\n",
			bubbleteautil.StyleForegroundSemanticError.Render(m.Err.Error()),
			bubbleteautil.StyleForegroundSemanticInfo.Render("Please hit enter to continue"),
		)
	}
	return b.String()
}

func (m SetupApp) ToResult() SetupAppResult {
	result := SetupAppResult{
		CertbotEnabled:                    m.mustFindQuestionByName(QuestionName_EnableCertbot).Value() == ValueTrue,
		AUTHGEAR_ONCE_ADMIN_USER_EMAIL:    m.mustFindQuestionByName(QuestionName_EnterAdminEmail).Value(),
		AUTHGEAR_ONCE_ADMIN_USER_PASSWORD: m.mustFindQuestionByName(QuestionName_EnterAdminPassword).Value(),
	}
	scheme := "http"
	if result.CertbotEnabled {
		scheme = "https"
		result.AUTHGEAR_CERTBOT_ENVIRONMENT = m.mustFindQuestionByName(QuestionName_SelectCertbotEnvironment).Value()
	}
	result.AUTHGEAR_HTTP_ORIGIN_PROJECT = fmt.Sprintf("%v://%v", scheme, m.mustFindQuestionByName(QuestionName_EnterDomain_Project).Value())
	result.AUTHGEAR_HTTP_ORIGIN_PORTAL = fmt.Sprintf("%v://%v", scheme, m.mustFindQuestionByName(QuestionName_EnterDomain_Portal).Value())
	result.AUTHGEAR_HTTP_ORIGIN_ACCOUNTS = fmt.Sprintf("%v://%v", scheme, m.mustFindQuestionByName(QuestionName_EnterDomain_Accounts).Value())

	switch m.mustFindQuestionByName(QuestionName_SelectSMTP).Value() {
	case SMTPSkip:
		break
	default:
		opts := m.makeSendTestEmailOptions()
		result.SMTPHost = opts.Host
		result.SMTPPort = opts.Port
		result.SMTPUsername = opts.Username
		result.SMTPPassword = opts.Password
		result.SMTPSenderAddress = opts.SenderAddress
	}
	return result
}

func newQuestion(q Question) Question {
	// q is copied.
	q.Model = q.Model.Focus()
	return q
}
