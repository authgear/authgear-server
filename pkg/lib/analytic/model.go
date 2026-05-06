package analytic

type AppCollaborator struct {
	AppID  string
	UserID string
}

type AppConfigSource struct {
	AppID    string
	Data     map[string][]byte
	PlanName string
}

type ReportData struct {
	Header []any
	Values [][]any
}

// DataPoint represent data point of a chart
type DataPoint struct {
	Label string `json:"label"`
	Data  int    `json:"data"`
}
