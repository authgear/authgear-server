package hook

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	MockNonBlockingEventType1 event.Type = "nonblockingevent.one"
	MockNonBlockingEventType2 event.Type = "nonblockingevent.two"
	MockNonBlockingEventType3 event.Type = "nonblockingevent.three"
	MockNonBlockingEventType4 event.Type = "nonblockingevent.four"

	MockBlockingEventType1 event.Type = "blockingevent.one"
	MockBlockingEventType2 event.Type = "blockingevent.two"
)

type MockUserEventBase struct {
	User model.User `json:"user"`
}

func (e *MockUserEventBase) UserID() string {
	return e.User.ID
}

func (e *MockUserEventBase) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

type MockNonBlockingEvent1 struct {
	MockUserEventBase
}

func (e *MockNonBlockingEvent1) NonBlockingEventType() event.Type {
	return MockNonBlockingEventType1
}

func (e *MockNonBlockingEvent1) FillContext(ctx *event.Context) {
}

func (e *MockNonBlockingEvent1) ForWebHook() bool {
	return true
}

func (e *MockNonBlockingEvent1) ForAudit() bool {
	return true
}

type MockNonBlockingEvent2 struct {
	MockUserEventBase
}

func (e *MockNonBlockingEvent2) NonBlockingEventType() event.Type {
	return MockNonBlockingEventType2
}

func (e *MockNonBlockingEvent2) FillContext(ctx *event.Context) {
}

func (e *MockNonBlockingEvent2) ForWebHook() bool {
	return true
}

func (e *MockNonBlockingEvent2) ForAudit() bool {
	return true
}

type MockNonBlockingEvent3 struct {
	MockUserEventBase
}

func (e *MockNonBlockingEvent3) FillContext(ctx *event.Context) {
}

func (e *MockNonBlockingEvent3) NonBlockingEventType() event.Type {
	return MockNonBlockingEventType3
}

func (e *MockNonBlockingEvent3) ForWebHook() bool {
	return true
}

func (e *MockNonBlockingEvent3) ForAudit() bool {
	return true
}

type MockNonBlockingEvent4 struct {
	MockUserEventBase
}

func (e *MockNonBlockingEvent4) NonBlockingEventType() event.Type {
	return MockNonBlockingEventType4
}

func (e *MockNonBlockingEvent4) FillContext(ctx *event.Context) {
}

func (e *MockNonBlockingEvent4) ForWebHook() bool {
	return true
}

func (e *MockNonBlockingEvent4) ForAudit() bool {
	return true
}

type MockBlockingEvent1 struct {
	MockUserEventBase
}

func (e *MockBlockingEvent1) BlockingEventType() event.Type {
	return MockBlockingEventType1
}

func (e *MockBlockingEvent1) FillContext(ctx *event.Context) {
}

func (e *MockBlockingEvent1) ApplyMutations(mutations event.Mutations) (event.BlockingPayload, bool) {
	user, mutated := ApplyMutations(e.User, mutations)
	if mutated {
		copied := *e
		copied.User = user
		return &copied, true
	}

	return e, false
}

func (e *MockBlockingEvent1) GenerateFullMutations() event.Mutations {
	return GenerateFullMutations(e.User)
}

type MockBlockingEvent2 struct {
	MockUserEventBase
}

func (e *MockBlockingEvent2) BlockingEventType() event.Type {
	return MockBlockingEventType2
}

func (e *MockBlockingEvent2) FillContext(ctx *event.Context) {
}

func (e *MockBlockingEvent2) ApplyMutations(mutations event.Mutations) (event.BlockingPayload, bool) {
	user, mutated := ApplyMutations(e.User, mutations)
	if mutated {
		copied := *e
		copied.User = user
		return &copied, true
	}

	return e, false
}

func (e *MockBlockingEvent2) GenerateFullMutations() (out event.Mutations) {
	return GenerateFullMutations(e.User)
}

var _ event.NonBlockingPayload = &MockNonBlockingEvent1{}
var _ event.NonBlockingPayload = &MockNonBlockingEvent2{}
var _ event.NonBlockingPayload = &MockNonBlockingEvent3{}
var _ event.NonBlockingPayload = &MockNonBlockingEvent4{}

var _ event.BlockingPayload = &MockBlockingEvent1{}
var _ event.BlockingPayload = &MockBlockingEvent2{}

func ApplyMutations(user model.User, mutations event.Mutations) (out model.User, mutated bool) {
	if mutations.User.StandardAttributes != nil {
		user.StandardAttributes = mutations.User.StandardAttributes
		mutated = true
	}
	if mutations.User.CustomAttributes != nil {
		user.CustomAttributes = mutations.User.CustomAttributes
		mutated = true
	}

	out = user
	return
}

func GenerateFullMutations(user model.User) event.Mutations {
	return event.Mutations{
		User: event.UserMutations{
			StandardAttributes: user.StandardAttributes,
			CustomAttributes:   user.CustomAttributes,
		},
	}
}
