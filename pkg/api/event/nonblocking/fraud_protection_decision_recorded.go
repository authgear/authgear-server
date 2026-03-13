package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	FraudProtectionDecisionRecorded event.Type = "fraud_protection.decision_recorded"
)

type FraudProtectionDecisionRecordedEventPayload struct {
	Record model.FraudProtectionDecisionRecord `json:"record"`
}

func (e *FraudProtectionDecisionRecordedEventPayload) NonBlockingEventType() event.Type {
	return FraudProtectionDecisionRecorded
}

func (e *FraudProtectionDecisionRecordedEventPayload) UserID() string {
	return e.Record.UserID
}

func (e *FraudProtectionDecisionRecordedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *FraudProtectionDecisionRecordedEventPayload) FillContext(ctx *event.Context) {}

func (e *FraudProtectionDecisionRecordedEventPayload) ForHook() bool {
	return false
}

func (e *FraudProtectionDecisionRecordedEventPayload) ForAudit() bool {
	return true
}

func (e *FraudProtectionDecisionRecordedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *FraudProtectionDecisionRecordedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &FraudProtectionDecisionRecordedEventPayload{}
