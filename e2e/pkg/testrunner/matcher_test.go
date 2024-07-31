package testrunner

import (
	"reflect"
	"testing"
)

func TestMatchJSON(t *testing.T) {
	type testCase struct {
		name     string
		jsonStr  string
		schema   string
		expected []MatchViolation
	}

	testCases := []testCase{
		{
			name: "Perfect Match",
			jsonStr: `{
				"id": 1,
				"title": "Test Article",
				"publish": true,
				"type": "articles",
				"tags": [],
				"error": null
			}`,
			schema: `{
				"id": "[[number]]",
				"title": "[[string]]",
				"publish": "[[boolean]]",
				"type": "articles",
				"tags": ["[[arrayof]]", "[[object]]"],
				"error": "[[null]]",
				"ignoreme": "[[ignore]]"
			}`,
			expected: []MatchViolation{},
		},
		{
			name:    "Type Mismatch",
			jsonStr: `{"id": "1", "title": "Test Article"}`,
			schema:  `{"id": "[[number]]", "title": "[[string]]"}`,
			expected: []MatchViolation{
				{
					Path:     "/id",
					Message:  "type mismatch",
					Expected: "[[number]]",
					Actual:   "[[string]]",
				},
			},
		},
		{
			name:    "Missing Field",
			jsonStr: `{"id": 1}`,
			schema:  `{"id": "[[number]]", "title": "[[string]]"}`,
			expected: []MatchViolation{
				{
					Path:     "/title",
					Message:  "missing field",
					Expected: "[[string]]",
					Actual:   "<missing>",
				},
			},
		},
		{
			name:     "Ignore Field",
			jsonStr:  `{"id": 1, "ignoreme": "ignore this"}`,
			schema:   `{"id": "[[number]]", "ignoreme": "[[ignore]]"}`,
			expected: []MatchViolation{},
		},
		{
			name:     "Nested Object",
			jsonStr:  `{"user": {"id": 1, "name": "John Doe"}}`,
			schema:   `{"user": {"id": "[[number]]", "name": "[[string]]"}}`,
			expected: []MatchViolation{},
		},
		{
			name:    "Array Mismatch",
			jsonStr: `{"tags": ["first", 2]}`,
			schema:  `{"tags": ["[[arrayof]]", "[[string]]"]}`,
			expected: []MatchViolation{
				{
					Path:     "/tags/1",
					Message:  "type mismatch",
					Expected: "[[string]]",
					Actual:   "[[number]]",
				},
			},
		},
		{
			name:    "Extra field",
			jsonStr: `{"id": 1, "title": "Test Article", "extra": "extra field"}`,
			schema:  `{"id": "[[number]]", "title": "[[string]]", "[[...rest]]": "[[number]]"}`,
			expected: []MatchViolation{
				{
					Path:     "/extra",
					Message:  "type mismatch",
					Expected: "[[number]]",
					Actual:   "[[string]]",
				},
			},
		},
		{
			name: "Nested object and array",
			jsonStr: `{
				"articles": [
					{"id": 1, "title": "Test Article 1"},
					{"id": 2, "title": "Test Article 2"}
				]
			}`,
			schema: `{
				"articles": ["[[arrayof]]", {"id": "[[number]]", "title": "[[string]]"}]
			}`,
			expected: []MatchViolation{},
		},
		{
			name: "Nested object and array mismatch",
			jsonStr: `{
				"articles": [
					{"id": 1, "title": "Test Article 1"},
					{"id": 2, "title": 2}
				]
			}`,
			schema: `{
				"articles": ["[[arrayof]]", {"id": "[[number]]", "title": "[[string]]"}]
			}`,
			expected: []MatchViolation{
				{
					Path:     "/articles/1/title",
					Message:  "type mismatch",
					Expected: "[[string]]",
					Actual:   "[[number]]",
				},
			},
		},
		{
			name: "Compare json with primtive",
			jsonStr: `{
				"id": 1
			}`,
			schema: `"[[number]]"`,
			expected: []MatchViolation{
				{
					Path:     "",
					Message:  "type mismatch",
					Expected: "[[number]]",
					Actual:   "[[object]]",
				},
			},
		},
		{
			name:     "Compare primitive",
			jsonStr:  `1`,
			schema:   `"[[number]]"`,
			expected: []MatchViolation{},
		},
		{
			name:    "Never field",
			jsonStr: `{"id": 1, "title": "Test Article", "extra": "extra field"}`,
			schema:  `{"id": "[[number]]", "title": "[[string]]", "[[...rest]]": "[[never]]"}`,
			expected: []MatchViolation{
				{
					Path:     "/extra",
					Message:  "type mismatch",
					Expected: "[[never]]",
					Actual:   "[[string]]",
				},
			},
		},
		{
			name: "Tuple match",
			jsonStr: `{
		    "tuple": [1, "string", true, null]
		  }`,
			schema: `{
		    "tuple": ["[[number]]", "[[string]]", "[[boolean]]", "[[null]]"]
		  }`,
			expected: []MatchViolation{},
		},
		{
			name: "Tuple violation",
			jsonStr: `{
		    "tuple": [1, "string", true, 1, null]
		  }`,
			schema: `{
		    "tuple": ["[[number]]", "[[string]]", "[[boolean]]", "[[null]]", "[[never]]"]
		  }`,
			expected: []MatchViolation{
				{
					Path:     "/tuple/3",
					Message:  "type mismatch",
					Expected: "[[null]]",
					Actual:   "[[number]]",
				},
				{
					Path:     "/tuple/4",
					Message:  "type mismatch",
					Expected: "[[never]]",
					Actual:   "[[null]]",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			violations, err := MatchJSON(tc.jsonStr, tc.schema)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(violations, tc.expected) {
				t.Errorf("Expected violations: %+v, got: %+v", tc.expected, violations)
			}
		})
	}
}
