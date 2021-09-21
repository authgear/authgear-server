package analytic

type AppCollaborator struct {
	AppID  string
	UserID string
}

type ReportData struct {
	Header []interface{}
	Values [][]interface{}
}

// DataPoint represent data point of a chart
type DataPoint struct {
	Label string `json:"label"`
	Data  int    `json:"data"`
}
