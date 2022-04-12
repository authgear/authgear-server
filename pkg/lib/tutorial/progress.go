package tutorial

type Progress string

const (
	ProgressAuthui            = "authui"
	ProgressCustomizeUI       = "customize_ui"
	ProgressCreateApplication = "create_application"
	ProgressSSO               = "sso"
	ProgressInvite            = "invite"
)

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
