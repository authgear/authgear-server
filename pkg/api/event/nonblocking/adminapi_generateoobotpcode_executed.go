package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AdminAPIGenerateOOBOTPCodeExecuted event.Type = "admin_api.generate_oob_otp_code.executed"
)

type AdminAPIGenerateOOBOTPCodeExecutedEventPayload struct {
	Target string `json:"target"`
}

func (e *AdminAPIGenerateOOBOTPCodeExecutedEventPayload) NonBlockingEventType() event.Type {
	return AdminAPIGenerateOOBOTPCodeExecuted
}

func (e *AdminAPIGenerateOOBOTPCodeExecutedEventPayload) UserID() string {
	return ""
}

func (e *AdminAPIGenerateOOBOTPCodeExecutedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeAdminAPI
}

func (e *AdminAPIGenerateOOBOTPCodeExecutedEventPayload) FillContext(ctx *event.Context) {
}

func (e *AdminAPIGenerateOOBOTPCodeExecutedEventPayload) ForHook() bool {
	return false
}

func (e *AdminAPIGenerateOOBOTPCodeExecutedEventPayload) ForAudit() bool {
	return true
}

func (e *AdminAPIGenerateOOBOTPCodeExecutedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *AdminAPIGenerateOOBOTPCodeExecutedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &AdminAPIGenerateOOBOTPCodeExecutedEventPayload{}
