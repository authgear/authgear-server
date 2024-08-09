package testrunner

import (
	"encoding/json"
	"fmt"
)

func MatchOutputQueryResult(output QueryOutput, rows []interface{}) (resultViolations []MatchViolation, err error) {
	if len(output.Rows) != 0 {
		if len(rows) == 0 {
			return nil, fmt.Errorf("expected query output, got empty query output")
		}

		rowsJSON, err := json.Marshal(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal rows: %w", err)
		}

		resultViolations, err = MatchJSON(string(rowsJSON), output.Rows)
		if err != nil {
			return nil, err
		}
	}

	return resultViolations, nil
}
