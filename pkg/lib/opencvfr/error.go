package opencvfr

import (
	"errors"
	"fmt"

	opencvfrapi "github.com/authgear/authgear-server/pkg/lib/opencvfr/api"
)

type Name string

const (
	NoMatchingFaceFound  Name = "NoMatchingFaceFound"
	SpoofedImageDetected Name = "SpoofedImageDetected"
	ServiceUnavailable   Name = "ServiceUnavailable"
	InvalidInput         Name = "InvalidInput"
	OperationNotAllowed  Name = "OperationNotAllowed"
	UnexpectedError      Name = "UnexpectedError"
)

type APIError struct {
	RawErr  *opencvfrapi.OpenCVFRAPIError
	Name    Name   `json:"name"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}

	if e.RawErr == nil {
		return fmt.Sprintf("opencvfr api error (%s): %s", e.Name, e.Message)
	}

	return fmt.Sprintf("opencvfr api error (%s [%s]): %s", e.Name, e.RawErr.OpenCVFRErrCode, e.Message)
}

func newNoMatchingFaceFoundError() *APIError {
	return &APIError{
		Name:    NoMatchingFaceFound,
		Message: "No matching face found",
	}
}

func newSpoofedImageDetectedError() *APIError {
	return &APIError{
		Name:    SpoofedImageDetected,
		Message: "Spoofed image detected.",
	}
}

func AsAPIError(err error) *APIError {
	if err == nil {
		return nil
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return nil
}
func getAPIError(err error) *APIError {
	if err == nil {
		return nil
	}

	var apiErr *opencvfrapi.OpenCVFRAPIError
	if errors.As(err, &apiErr) {
		return constructAPIError(apiErr)
	}

	return nil
}

// constructAPIError constructs an APIError from an OpenCVFRAPIError
func constructAPIError(err *opencvfrapi.OpenCVFRAPIError) *APIError {
	n := asName(err.ErrCode())
	if n == "" {
		return nil
	}
	return &APIError{
		RawErr:  err,
		Name:    n,
		Message: err.Message,
	}
}

func asName(opencvfrErrCode opencvfrapi.ErrCode) Name {
	switch opencvfrErrCode {
	// ServiceUnavailable
	case opencvfrapi.MaintenanceMode:
		return ServiceUnavailable
	// InvalidInput
	case opencvfrapi.FaceTooSmall:
		fallthrough
	case opencvfrapi.NoFacesFound:
		fallthrough
	case opencvfrapi.FaceRotation:
		fallthrough
	case opencvfrapi.FaceEdgesNotVisible:
		fallthrough
	case opencvfrapi.FaceOccluded:
		fallthrough
	case opencvfrapi.FaceTooClose:
		fallthrough
	case opencvfrapi.FaceCropped:
		fallthrough
	case opencvfrapi.InvalidFaceForLiveness:
		fallthrough
	case opencvfrapi.MultipleFaces:
		fallthrough
	case opencvfrapi.BlurryImage:
		return InvalidInput
	// NoMatchingFaceFound
	case opencvfrapi.EntityNotFound:
		return NoMatchingFaceFound
	// OperationNotAllowed
	case opencvfrapi.DuplicateEntity:
		fallthrough
	case opencvfrapi.PersonLimitExceeded:
		fallthrough
	case opencvfrapi.UpdateResultsInTooManyThumbnails:
		fallthrough
	case opencvfrapi.UpdateResultsInNoThumbnail:
		fallthrough
	case opencvfrapi.DeleteResultsInNoThumbnail:
		fallthrough
	case opencvfrapi.NotStandardCrop: // TODO (identity-week-demo): Confirm when does this error occur
		return OperationNotAllowed
		// UnexpectedError
	case opencvfrapi.ImageCouldNotBeDecoded:
		fallthrough
	case opencvfrapi.RateLimitExceeded:
		fallthrough
	case opencvfrapi.InvalidAPIKey:
		fallthrough
	case opencvfrapi.InvalidToken:
		fallthrough
	case opencvfrapi.ExpiredSubscription:
		fallthrough
	case opencvfrapi.MasterKeyCannotBeUsed:
		fallthrough
	case opencvfrapi.DisabledAPIKey:
		fallthrough
	case opencvfrapi.RestrictedAPI:
		fallthrough
	case opencvfrapi.ReadOnlyAPIKey:
		fallthrough
	default:
		return UnexpectedError
	}
}
