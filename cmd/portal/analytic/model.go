package analytic

const ReportOutputTypeStdout = "stdout"
const ReportOutputTypeGoogleSheets = "google-sheets"

type ReportData struct {
	Header []interface{}
	Values [][]interface{}
}
