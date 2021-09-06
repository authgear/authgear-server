package analytic

type AppCollaborator struct {
	AppID  string
	UserID string
}

type ReportData struct {
	Header []interface{}
	Values [][]interface{}
}
