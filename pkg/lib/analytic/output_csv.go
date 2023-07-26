package analytic

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
)

type OutputCSV struct {
	Writer io.Writer
}

func (o *OutputCSV) OutputReport(ctx context.Context, data *ReportData) error {
	w := csv.NewWriter(o.Writer)
	defer w.Flush()
	csvData, err := o.convert(data)
	if err != nil {
		return fmt.Errorf("unable to convert to csv data: %w", err)
	}

	err = w.WriteAll(csvData)
	if err != nil {
		return fmt.Errorf("unable to write csv file: %w", err)
	}

	return nil
}

func (*OutputCSV) convert(data *ReportData) (records [][]string, err error) {
	toString := func(i interface{}) (string, error) {
		switch i := i.(type) {
		case int:
			return fmt.Sprintf("%d", i), nil
		case string:
			return i, nil
		default:
			// support convert more types if needed
			return "", fmt.Errorf("unknown data type: %T", i)
		}
	}
	// perpare the header
	row := make([]string, len(data.Header))
	for i, d := range data.Header {
		row[i], err = toString(d)
		if err != nil {
			return
		}
	}
	records = append(records, row)

	// perpare the data
	for _, valRow := range data.Values {
		row := make([]string, len(valRow))
		for i, d := range valRow {
			row[i], err = toString(d)
			if err != nil {
				return
			}
		}
		records = append(records, row)
	}

	return
}
