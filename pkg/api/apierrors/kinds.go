package apierrors

import "net/http"

type Name string

const (
	BadRequest            Name = "BadRequest"
	Invalid               Name = "Invalid"
	Unauthorized          Name = "Unauthorized"
	Forbidden             Name = "Forbidden"
	NotFound              Name = "NotFound"
	AlreadyExists         Name = "AlreadyExists"
	DataRace              Name = "DataRace"
	TooManyRequest        Name = "TooManyRequest"
	InternalError         Name = "InternalError"
	ServiceUnavailable    Name = "ServiceUnavailable"
	RequestEntityTooLarge Name = "RequestEntityTooLarge"
)

func (n Name) HTTPStatus() int {
	switch n {
	case BadRequest, Invalid:
		return http.StatusBadRequest
	case Unauthorized:
		return http.StatusUnauthorized
	case Forbidden:
		return http.StatusForbidden
	case NotFound:
		return http.StatusNotFound
	case AlreadyExists, DataRace:
		return http.StatusConflict
	case TooManyRequest:
		return http.StatusTooManyRequests
	case InternalError:
		return http.StatusInternalServerError
	case ServiceUnavailable:
		return http.StatusServiceUnavailable
	case RequestEntityTooLarge:
		return http.StatusRequestEntityTooLarge
	default:
		return http.StatusInternalServerError
	}
}

type Kind struct {
	Name   Name   `json:"name"`
	Reason string `json:"reason"`

	IsSkipLoggingToExternalService bool `json:"-"`
}

func (k Kind) SkipLoggingToExternalService() Kind {
	kk := k
	kk.IsSkipLoggingToExternalService = true
	return kk
}

func (n Name) WithReason(reason string) Kind {
	return Kind{Name: n, Reason: reason}
}

var ValidationFailed = Invalid.WithReason("ValidationFailed")
