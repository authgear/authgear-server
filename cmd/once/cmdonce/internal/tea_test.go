package internal

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIndentLines(t *testing.T) {
	Convey("indentLines", t, func() {
		cases := []struct {
			name     string
			input    string
			indent   string
			expected string
		}{
			{
				name:     "empty string",
				input:    ``,
				indent:   ``,
				expected: ``,
			},
			{
				name: "single line",
				input: `This is a sentence.
`,
				indent: `  `,
				expected: `  This is a sentence.
`,
			},
			{
				name: "multi lines",
				input: `sentence 1
sentence 2
`,
				indent: `  `,
				expected: `  sentence 1
  sentence 2
`,
			},
			{
				name: "original indentation is kept",
				input: `title
  - list item 1
  - list item 2
`,
				indent: `  `,
				expected: `  title
    - list item 1
    - list item 2
`,
			},
		}

		for _, c := range cases {
			Convey(c.name, func() {
				actual := indentLines(c.input, c.indent)
				So(actual, ShouldEqual, c.expected)
			})
		}
	})
}
