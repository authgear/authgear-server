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

type MockNonBlockingEvent1 struct {
	MockUserEventBase
}

func (e *MockNonBlockingEvent1) NonBlockingEventType() event.Type {
	return MockNonBlockingEventType1
}

type MockNonBlockingEvent2 struct {
	MockUserEventBase
}

func (e *MockNonBlockingEvent2) NonBlockingEventType() event.Type {
	return MockNonBlockingEventType2
}

type MockNonBlockingEvent3 struct {
	MockUserEventBase
}

func (e *MockNonBlockingEvent3) NonBlockingEventType() event.Type {
	return MockNonBlockingEventType3
}

type MockNonBlockingEvent4 struct {
	MockUserEventBase
}

func (e *MockNonBlockingEvent4) NonBlockingEventType() event.Type {
	return MockNonBlockingEventType4
}

type MockBlockingEvent1 struct {
	MockUserEventBase
}

func (e *MockBlockingEvent1) BlockingEventType() event.Type {
	return MockBlockingEventType1
}

type MockBlockingEvent2 struct {
	MockUserEventBase
}

func (e *MockBlockingEvent2) BlockingEventType() event.Type {
	return MockBlockingEventType2
}

var _ event.NonBlockingPayload = &MockNonBlockingEvent1{}
var _ event.NonBlockingPayload = &MockNonBlockingEvent2{}
var _ event.NonBlockingPayload = &MockNonBlockingEvent3{}
var _ event.NonBlockingPayload = &MockNonBlockingEvent4{}

var _ event.BlockingPayload = &MockBlockingEvent1{}
var _ event.BlockingPayload = &MockBlockingEvent2{}
