package testrunner

import (
	"reflect"
	"testing"
)

func TestSortHookServerRowsBySeq(t *testing.T) {
	row := func(seq interface{}, name string) map[string]interface{} {
		return map[string]interface{}{"seq": seq, "name": name}
	}

	type testCase struct {
		name     string
		rows     []interface{}
		expected []interface{}
	}

	testCases := []testCase{
		{
			name:     "Empty",
			rows:     []interface{}{},
			expected: []interface{}{},
		},
		{
			name: "Out of order rows are sorted by seq",
			rows: []interface{}{
				row(float64(470), "block"),
				row(float64(468), "alert"),
			},
			expected: []interface{}{
				row(float64(468), "alert"),
				row(float64(470), "block"),
			},
		},
		{
			name: "Already ordered rows are unchanged",
			rows: []interface{}{
				row(float64(468), "alert"),
				row(float64(470), "block"),
			},
			expected: []interface{}{
				row(float64(468), "alert"),
				row(float64(470), "block"),
			},
		},
		{
			name: "Equal seq keeps arrival order",
			rows: []interface{}{
				row(float64(468), "second"),
				row(float64(468), "first"),
			},
			expected: []interface{}{
				row(float64(468), "second"),
				row(float64(468), "first"),
			},
		},
		{
			name: "Rows without seq keep arrival order",
			rows: []interface{}{
				row(float64(470), "block"),
				map[string]interface{}{"name": "no-seq"},
				row(float64(468), "alert"),
			},
			expected: []interface{}{
				row(float64(470), "block"),
				map[string]interface{}{"name": "no-seq"},
				row(float64(468), "alert"),
			},
		},
		{
			name: "Non-object rows keep arrival order",
			rows: []interface{}{
				row(float64(470), "block"),
				"not an object",
				row(float64(468), "alert"),
			},
			expected: []interface{}{
				row(float64(470), "block"),
				"not an object",
				row(float64(468), "alert"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sortHookServerRowsBySeq(tc.rows)
			if !reflect.DeepEqual(tc.rows, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, tc.rows)
			}
		})
	}
}
