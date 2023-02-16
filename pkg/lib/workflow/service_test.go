package workflow

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/clientid"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	RegisterPrivateIntent(&intentServiceContext{})
	RegisterPrivateIntent(&intentNilInput{})
	RegisterNode(&nodeServiceContext{})
	RegisterNode(&nodeNilInput{})
}

func TestService(t *testing.T) {
	Convey("Service", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.TODO()
		deps := &Dependencies{}
		logger := ServiceLogger{log.Null}
		savepoint := NewMockSavepoint(ctrl)
		store := NewMockStore(ctrl)

		service := &Service{
			ContextDoNotUseDirectly: ctx,
			Deps:                    deps,
			Logger:                  logger,
			Store:                   store,
			Savepoint:               savepoint,
		}

		Convey("CreateNewWorkflow with intent expecting non-nil input at the beginning", func() {
			rng = rand.New(rand.NewSource(0))

			intent := &intentAuthenticate{
				PretendLoginIDExists: false,
			}

			gomock.InOrder(
				store.EXPECT().CreateSession(gomock.Any()).Return(nil),
				savepoint.EXPECT().Begin().Times(1).Return(nil),
				store.EXPECT().CreateWorkflow(gomock.Any()).Return(nil),
				savepoint.EXPECT().Rollback().Times(1).Return(nil),
			)

			output, err := service.CreateNewWorkflow(intent, &SessionOptions{})
			So(err, ShouldBeNil)
			So(output, ShouldResemble, &ServiceOutput{
				Action: &WorkflowAction{
					Type: WorkflowActionTypeContinue,
				},
				Workflow: &Workflow{
					WorkflowID: "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					InstanceID: "1WPH8EXJFWMAZ7M8Y9EGAG34SPW86VXT",
					Intent:     intent,
				},
				WorkflowOutput: &WorkflowOutput{
					WorkflowID: "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					InstanceID: "1WPH8EXJFWMAZ7M8Y9EGAG34SPW86VXT",
					Intent: IntentOutput{
						Kind: "intentAuthenticate",
						Data: nil,
					},
				},
				Session: &Session{
					WorkflowID: "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
				},
				SessionOutput: &SessionOutput{
					WorkflowID: "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
				},
			})
		})

		Convey("CreateNewWorkflow with intent expecting nil input at the beginning", func() {
			rng = rand.New(rand.NewSource(0))

			intent := &intentNilInput{}

			gomock.InOrder(
				store.EXPECT().CreateSession(gomock.Any()).Return(nil),
				savepoint.EXPECT().Begin().Times(1).Return(nil),
				store.EXPECT().CreateWorkflow(gomock.Any()).Return(nil),
				savepoint.EXPECT().Rollback().Times(1).Return(nil),
				savepoint.EXPECT().Begin().Times(1).Return(nil),
				savepoint.EXPECT().Commit().Times(1).Return(nil),
				store.EXPECT().DeleteSession(gomock.Any()).Return(nil),
				store.EXPECT().DeleteWorkflow(gomock.Any()).Return(nil),
			)

			output, err := service.CreateNewWorkflow(intent, &SessionOptions{})
			So(errors.Is(err, ErrEOF), ShouldBeTrue)
			So(output, ShouldResemble, &ServiceOutput{
				Action: &WorkflowAction{
					Type: WorkflowActionTypeFinish,
				},
				Workflow: &Workflow{
					WorkflowID: "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					InstanceID: "Y37GSHFPM7259WFBY64B4HTJ4PM8G482",
					Intent:     intent,
					Nodes: []Node{
						{
							Type:   NodeTypeSimple,
							Simple: &nodeNilInput{},
						},
					},
				},
				WorkflowOutput: &WorkflowOutput{
					WorkflowID: "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					InstanceID: "Y37GSHFPM7259WFBY64B4HTJ4PM8G482",
					Intent: IntentOutput{
						Kind: "intentNilInput",
						Data: nil,
					},
					Nodes: []NodeOutput{
						{
							Type: NodeTypeSimple,
							Simple: &NodeSimpleOutput{
								Kind: "nodeNilInput",
								Data: nil,
							},
						},
					},
				},
				Session: &Session{
					WorkflowID: "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
				},
				SessionOutput: &SessionOutput{
					WorkflowID: "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
				},
			})
		})

		Convey("FeedInput", func() {
			rng = rand.New(rand.NewSource(0))

			intent := &intentAuthenticate{
				PretendLoginIDExists: false,
			}
			workflow := &Workflow{
				WorkflowID: "workflow-id",
				InstanceID: "instance-id",
				Intent:     intent,
			}
			session := &Session{
				WorkflowID: "workflow-id",
			}

			gomock.InOrder(
				store.EXPECT().GetWorkflowByInstanceID(workflow.InstanceID).Times(1).Return(workflow, nil),
				store.EXPECT().GetSession(workflow.WorkflowID).Return(session, nil),
				savepoint.EXPECT().Begin().Times(1).Return(nil),
				store.EXPECT().GetWorkflowByInstanceID(workflow.InstanceID).Times(1).Return(workflow, nil),
				store.EXPECT().CreateWorkflow(gomock.Any()).Return(nil),
				savepoint.EXPECT().Rollback().Times(1).Return(nil),
			)

			output, err := service.FeedInput(workflow.WorkflowID, workflow.InstanceID, &inputLoginID{
				LoginID: "user@example.com",
			})
			So(err, ShouldBeNil)
			So(output, ShouldResemble, &ServiceOutput{
				Action: &WorkflowAction{
					Type: WorkflowActionTypeContinue,
				},
				Workflow: &Workflow{
					WorkflowID: "workflow-id",
					InstanceID: "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					Intent:     intent,
					Nodes: []Node{
						{
							Type: NodeTypeSubWorkflow,
							SubWorkflow: &Workflow{
								Intent: &intentSignup{
									LoginID: "user@example.com",
								},
								Nodes: []Node{
									{
										Type: NodeTypeSubWorkflow,

										SubWorkflow: &Workflow{
											Intent: &intentAddLoginID{
												LoginID: "user@example.com",
											},
											Nodes: []Node{
												{
													Type: NodeTypeSimple,
													Simple: &nodeVerifyLoginID{
														LoginID: "user@example.com",
														OTP:     "123456",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				WorkflowOutput: &WorkflowOutput{
					WorkflowID: "workflow-id",
					InstanceID: "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					Intent: IntentOutput{
						Kind: "intentAuthenticate",
						Data: nil,
					},
					Nodes: []NodeOutput{
						{
							Type: NodeTypeSubWorkflow,
							SubWorkflow: &WorkflowOutput{
								Intent: IntentOutput{
									Kind: "intentSignup",
									Data: nil,
								},
								Nodes: []NodeOutput{
									{
										Type: NodeTypeSubWorkflow,
										SubWorkflow: &WorkflowOutput{
											Intent: IntentOutput{
												Kind: "intentAddLoginID",
												Data: nil,
											},
											Nodes: []NodeOutput{
												{
													Type: NodeTypeSimple,
													Simple: &NodeSimpleOutput{
														Kind: "nodeVerifyLoginID",
														Data: nil,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Session: &Session{
					WorkflowID: "workflow-id",
				},
				SessionOutput: &SessionOutput{
					WorkflowID: "workflow-id",
				},
			})
		})
	})
}

type intentNilInput struct{}

func (*intentNilInput) Kind() string {
	return "intentNilInput"
}

func (*intentNilInput) JSONSchema() *validation.SimpleSchema {
	return EmptyJSONSchema
}

func (*intentNilInput) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (*intentNilInput) CanReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Input, error) {
	if len(workflow.Nodes) == 0 {
		return nil, nil
	}
	return nil, ErrEOF
}

func (*intentNilInput) ReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
	return NewNodeSimple(&nodeNilInput{}), nil
}

func (*intentNilInput) OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error) {
	return nil, nil
}

type nodeNilInput struct {
	ClientID string
}

func (*nodeNilInput) Kind() string {
	return "nodeNilInput"
}

func (*nodeNilInput) GetEffects(ctx context.Context, deps *Dependencies, w *Workflow) ([]Effect, error) {
	return nil, nil
}

func (*nodeNilInput) CanReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Input, error) {
	return nil, ErrEOF
}

func (*nodeNilInput) ReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
	return nil, ErrIncompatibleInput
}

func (*nodeNilInput) OutputData(ctx context.Context, deps *Dependencies, w *Workflow) (interface{}, error) {
	return nil, nil
}

type intentServiceContext struct{}

var _ CookieGetter = &intentServiceContext{}

func (*intentServiceContext) Kind() string {
	return "intentServiceContext"
}

func (*intentServiceContext) JSONSchema() *validation.SimpleSchema {
	return EmptyJSONSchema
}

func (*intentServiceContext) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (*intentServiceContext) CanReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Input, error) {
	if len(workflow.Nodes) == 0 {
		return []Input{
			&inputServiceContext{},
		}, nil
	}
	return nil, ErrEOF
}

func (*intentServiceContext) ReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
	var inputServiceContext InputServiceContext

	switch {
	case AsInput(input, &inputServiceContext):
		return NewNodeSimple(&nodeServiceContext{
			ClientID: clientid.GetClientID(ctx),
		}), nil
	default:
		return nil, ErrIncompatibleInput
	}

}

func (*intentServiceContext) OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error) {
	return nil, nil
}

func (*intentServiceContext) GetCookies(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]*http.Cookie, error) {
	return []*http.Cookie{
		{
			Name:  "intentServiceContext",
			Value: "intentServiceContext",
		},
	}, nil
}

type inputServiceContext struct{}

func (*inputServiceContext) Kind() string {
	return "inputServiceContext"
}

func (*inputServiceContext) JSONSchema() *validation.SimpleSchema {
	return EmptyJSONSchema
}

func (*inputServiceContext) Marker() {}

type InputServiceContext interface {
	Marker()
}

type nodeServiceContext struct {
	ClientID string
}

var _ CookieGetter = &nodeServiceContext{}

func (*nodeServiceContext) Kind() string {
	return "nodeServiceContext"
}

func (*nodeServiceContext) GetEffects(ctx context.Context, deps *Dependencies, w *Workflow) ([]Effect, error) {
	return nil, nil
}

func (*nodeServiceContext) CanReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Input, error) {
	return nil, ErrEOF
}

func (*nodeServiceContext) ReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
	return nil, ErrIncompatibleInput
}

func (*nodeServiceContext) OutputData(ctx context.Context, deps *Dependencies, w *Workflow) (interface{}, error) {
	return nil, nil
}

func (*nodeServiceContext) GetCookies(ctx context.Context, deps *Dependencies, w *Workflow) ([]*http.Cookie, error) {
	return []*http.Cookie{
		{
			Name:  "nodeServiceContext",
			Value: "nodeServiceContext",
		},
	}, nil
}

func TestServiceContext(t *testing.T) {
	Convey("Service Context", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.TODO()
		deps := &Dependencies{}
		logger := ServiceLogger{log.Null}
		savepoint := NewMockSavepoint(ctrl)
		store := NewMockStore(ctrl)

		service := &Service{
			ContextDoNotUseDirectly: ctx,
			Deps:                    deps,
			Logger:                  logger,
			Store:                   store,
			Savepoint:               savepoint,
		}

		Convey("Populate context", func() {
			rng = rand.New(rand.NewSource(0))

			intent := &intentServiceContext{}

			savepoint.EXPECT().Begin().AnyTimes().Return(nil)
			savepoint.EXPECT().Rollback().AnyTimes().Return(nil)
			savepoint.EXPECT().Commit().AnyTimes().Return(nil)
			store.EXPECT().CreateSession(gomock.Any()).Return(nil)
			store.EXPECT().CreateWorkflow(gomock.Any()).AnyTimes().Return(nil)
			store.EXPECT().DeleteSession(gomock.Any()).Return(nil)
			store.EXPECT().DeleteWorkflow(gomock.Any()).Return(nil)

			output, err := service.CreateNewWorkflow(intent, &SessionOptions{
				ClientID: "client-id",
			})
			So(err, ShouldBeNil)

			store.EXPECT().GetSession(output.Workflow.WorkflowID).Return(output.Session, nil)
			store.EXPECT().GetWorkflowByInstanceID(output.Workflow.InstanceID).Times(2).Return(output.Workflow, nil)

			output, err = service.FeedInput(
				output.Session.WorkflowID,
				output.Workflow.InstanceID,
				&inputServiceContext{},
			)
			So(errors.Is(err, ErrEOF), ShouldBeTrue)
			So(output, ShouldResemble, &ServiceOutput{
				Action: &WorkflowAction{
					Type: WorkflowActionTypeFinish,
				},
				Workflow: &Workflow{
					WorkflowID: "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					InstanceID: "Y37GSHFPM7259WFBY64B4HTJ4PM8G482",
					Intent:     intent,
					Nodes: []Node{
						{
							Type: NodeTypeSimple,
							Simple: &nodeServiceContext{
								ClientID: "client-id",
							},
						},
					},
				},
				WorkflowOutput: &WorkflowOutput{
					WorkflowID: "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					InstanceID: "Y37GSHFPM7259WFBY64B4HTJ4PM8G482",
					Intent: IntentOutput{
						Kind: "intentServiceContext",
						Data: nil,
					},
					Nodes: []NodeOutput{
						{
							Type: NodeTypeSimple,
							Simple: &NodeSimpleOutput{
								Kind: "nodeServiceContext",
								Data: nil,
							},
						},
					},
				},
				Session: &Session{
					WorkflowID: "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					ClientID:   "client-id",
				},
				SessionOutput: &SessionOutput{
					WorkflowID: "TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					ClientID:   "client-id",
				},
				Cookies: []*http.Cookie{
					{
						Name:  "nodeServiceContext",
						Value: "nodeServiceContext",
					},
					{
						Name:  "intentServiceContext",
						Value: "intentServiceContext",
					},
				},
			})
		})
	})
}
