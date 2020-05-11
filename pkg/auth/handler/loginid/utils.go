package loginid

import (
	"regexp"

	"github.com/skygeario/skygear-server/pkg/core/validation"
)

var loginIDPointerPrefixRegex = regexp.MustCompile(`^/(\d+)/`)

// correctErrorCausePointer check and update the error causes pointer
// with updatePointerFunc function
// updatePointerFunc provides the relative path of pointer and expect to return
// the corrected json pointer
func correctErrorCausePointer(err error, updatePointerFunc func(string) string) error {
	if causes := validation.ErrorCauses(err); len(causes) > 0 {
		for i, cause := range causes {
			matches := loginIDPointerPrefixRegex.FindStringSubmatch(cause.Pointer)
			if len(matches) > 0 {
				cause.Pointer = updatePointerFunc(cause.Pointer[len(matches[0]):])
			}
			causes[i] = cause
		}
		err = validation.NewValidationFailed("invalid login ID", causes)
	}
	return err
}
