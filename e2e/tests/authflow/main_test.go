package tests

import (
	"testing"
)

func TestAuthflow(t *testing.T) {
	testCases, err := loadAllTestCases("..")
	if err != nil {
		t.Errorf("error: %v", err)
		return
	}

	hasFocus := false
	for _, testCase := range testCases {
		if testCase.Focus {
			hasFocus = true
			break
		}
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.FullName(), func(t *testing.T) {
			t.Parallel()

			if hasFocus && !tc.Focus {
				t.Skip("skipping non-focus test case")
				return
			}

			runTestCase(t, tc)
		})
	}
}
