package api

import (
	"fmt"
)

type OpenCVFRAPIError struct {
	HTTPStatusCode  int    `json:"httpStatusCode"`
	Message         string `json:"message"`
	OpenCVFRErrCode string `json:"openCVFRErrCode"`
	RetryAfter      int    `json:"retryAfter"`
}

func (e *OpenCVFRAPIError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("opencvfr api error (%s): %s", e.OpenCVFRErrCode, e.Message)
}

type OpenCVFRValidationError struct {
	Details string `json:"details"`
}

func (e *OpenCVFRValidationError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("opencvfr validation error: %s", e.Details)
}
