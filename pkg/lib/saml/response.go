package saml

import (
	"time"

	crewjamsaml "github.com/crewjam/saml"
)

type Response struct {
	crewjamsaml.Response
}

func newResponse(issueInstant time.Time, status Status) *Response {
	return &Response{
		crewjamsaml.Response{
			ID:           GenerateID(),
			IssueInstant: issueInstant,
			Version:      SAMLVersion2,
			Status:       status.status,
		},
	}
}

func NewRequestDeniedErrorResponse(issueInstant time.Time, message string) *Response {
	var messageEl *crewjamsaml.StatusMessage
	if message != "" {
		messageEl = &crewjamsaml.StatusMessage{
			Value: message,
		}
	}
	status := Status{
		status: crewjamsaml.Status{
			StatusCode: crewjamsaml.StatusCode{
				Value: crewjamsaml.StatusRequester,
				StatusCode: &crewjamsaml.StatusCode{
					Value: crewjamsaml.StatusRequestDenied,
				},
			},
			StatusMessage: messageEl,
		},
	}
	return newResponse(issueInstant, status)
}

type Status struct {
	status crewjamsaml.Status
}
