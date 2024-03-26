package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIMutationGenerateOOBOTPCodeExecuted event.Type = "admin_api.mutation.generate_oob_otp_code.executed"
)

type AdminAPIMutationGenerateOOBOTPCodeExecutedEventPayload struct {
	Target  string `json:"target"`
	Purpose string `json:"purpose"`
}

func (e *AdminAPIMutationGenerateOOBOTPCodeExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIMutationGenerateOOBOTPCodeExecuted
}

func (e *AdminAPIMutationGenerateOOBOTPCodeExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIMutationGenerateOOBOTPCodeExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIMutationGenerateOOBOTPCodeExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIMutationGenerateOOBOTPCodeExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIMutationGenerateOOBOTPCodeExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIMutationGenerateOOBOTPCodeExecutedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AdminAPIMutationGenerateOOBOTPCodeExecutedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AdminAPIMutationGenerateOOBOTPCodeExecutedEventPayload{}
