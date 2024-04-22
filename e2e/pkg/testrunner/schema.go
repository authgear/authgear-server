package testrunner

import "github.com/authgear/authgear-server/pkg/util/validation"

var TestCaseSchema = validation.NewMultipartSchema("TestCase")

func init() {
	TestCaseSchema.Instantiate()
}

func DumpSchema() (string, error) {
	return TestCaseSchema.DumpSchemaString(true)
}
