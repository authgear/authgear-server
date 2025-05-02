package tutorial

type Progress string

const (
	ProgressAuthui            = "authui"
	ProgressCustomizeUI       = "customize_ui"
	ProgressCreateApplication = "create_application"
	ProgressSSO               = "sso"
	ProgressInvite            = "invite"
)

func ProgressFromString(s string) (Progress, bool) {
	switch Progress(s) {
	case ProgressAuthui:
		return ProgressAuthui, true
	case ProgressCustomizeUI:
		return ProgressCustomizeUI, true
	case ProgressCreateApplication:
		return ProgressCreateApplication, true
	case ProgressSSO:
		return ProgressSSO, true
	case ProgressInvite:
		return ProgressInvite, true
	default:
		return "", false
	}
}

type Entry struct {
	AppID string
	Data  map[string]interface{}
}

func NewEntry(appID string) *Entry {
	return &Entry{
		AppID: appID,
		Data: map[string]interface{}{
			"progress": make(map[string]interface{}),
			"skipped":  false,
		},
	}
}

func (e *Entry) AddProgress(ps []Progress) {
	m := e.Data["progress"].(map[string]interface{})
	for _, p := range ps {
		m[string(p)] = true
	}
}

func (e *Entry) Skip() {
	e.Data["skipped"] = true
}

func (e *Entry) SetProjectWizardData(data interface{}) {
	e.Data["project_wizard"] = data
}
