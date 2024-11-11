package workflow

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
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

		ctx := context.Background()
		deps := &Dependencies{}
		logger := ServiceLogger{log.Null}
		database := &db.MockHandle{}
		store := NewMockStore(ctrl)
		uiInfoResolver := NewMockServiceUIInfoResolver(ctrl)

		service := &Service{
			Deps:           deps,
			Logger:         logger,
			Store:          store,
			Database:       database,
			UIInfoResolver: uiInfoResolver,
		}

		Convey("CreateNewWorkflow with intent expecting non-nil input at the beginning", func() {
			rng = rand.New(rand.NewSource(0))

			intent := &intentAuthenticate{
				PretendLoginIDExists: false,
			}

			gomock.InOrder(
				store.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Return(nil),
				store.EXPECT().CreateWorkflow(gomock.Any(), gomock.Any()).Return(nil),
			)

			output, err := service.CreateNewWorkflow(ctx, intent, &SessionOptions{})
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

		SkipConvey("CreateNewWorkflow with intent expecting nil input at the beginning", func() {
			rng = rand.New(rand.NewSource(0))

			intent := &intentNilInput{}

			gomock.InOrder(
				store.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Return(nil),
				store.EXPECT().CreateWorkflow(gomock.Any(), gomock.Any()).Return(nil),

				uiInfoResolver.EXPECT().SetAuthenticationInfoInQuery(gomock.Any(), gomock.Any()).Return(""),

				store.EXPECT().DeleteSession(gomock.Any(), gomock.Any()).Return(nil),
				store.EXPECT().DeleteWorkflow(gomock.Any(), gomock.Any()).Return(nil),
			)

			output, err := service.CreateNewWorkflow(ctx, intent, &SessionOptions{})
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

		SkipConvey("FeedInput", func() {
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
				store.EXPECT().GetWorkflowByInstanceID(gomock.Any(), workflow.InstanceID).Times(1).Return(workflow, nil),
				store.EXPECT().GetSession(gomock.Any(), workflow.WorkflowID).Return(session, nil),
				store.EXPECT().GetWorkflowByInstanceID(gomock.Any(), workflow.InstanceID).Times(1).Return(workflow, nil),
				store.EXPECT().CreateWorkflow(gomock.Any(), gomock.Any()).Return(nil),
			)

			output, err := service.FeedInput(ctx, workflow.WorkflowID, workflow.InstanceID, "", &inputLoginID{
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

func (*intentNilInput) GetEffects(ctx context.Context, deps *Dependencies, workflows Workflows) ([]Effect, error) {
	return nil, nil
}

func (*intentNilInput) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) ([]Input, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return nil, nil
	}
	return nil, ErrEOF
}

func (*intentNilInput) ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error) {
	return NewNodeSimple(&nodeNilInput{}), nil
}

func (*intentNilInput) OutputData(ctx context.Context, deps *Dependencies, workflows Workflows) (interface{}, error) {
	return nil, nil
}

type nodeNilInput struct {
	ClientID string
}

func (*nodeNilInput) Kind() string {
	return "nodeNilInput"
}

func (*nodeNilInput) GetEffects(ctx context.Context, deps *Dependencies, workflows Workflows) ([]Effect, error) {
	return nil, nil
}

func (*nodeNilInput) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) ([]Input, error) {
	return nil, ErrEOF
}

func (*nodeNilInput) ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error) {
	return nil, ErrIncompatibleInput
}

func (*nodeNilInput) OutputData(ctx context.Context, deps *Dependencies, workflows Workflows) (interface{}, error) {
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

func (*intentServiceContext) GetEffects(ctx context.Context, deps *Dependencies, workflows Workflows) ([]Effect, error) {
	return nil, nil
}

func (*intentServiceContext) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) ([]Input, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return []Input{
			&inputServiceContext{},
		}, nil
	}
	return nil, ErrEOF
}

func (*intentServiceContext) ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error) {
	var inputServiceContext InputServiceContext

	switch {
	case AsInput(input, &inputServiceContext):
		return NewNodeSimple(&nodeServiceContext{
			ClientID: uiparam.GetUIParam(ctx).ClientID,
		}), nil
	default:
		return nil, ErrIncompatibleInput
	}

}

func (*intentServiceContext) OutputData(ctx context.Context, deps *Dependencies, workflows Workflows) (interface{}, error) {
	return nil, nil
}

func (*intentServiceContext) GetCookies(ctx context.Context, deps *Dependencies, workflows Workflows) ([]*http.Cookie, error) {
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

func (*nodeServiceContext) GetEffects(ctx context.Context, deps *Dependencies, workflows Workflows) ([]Effect, error) {
	return nil, nil
}

func (*nodeServiceContext) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) ([]Input, error) {
	return nil, ErrEOF
}

func (*nodeServiceContext) ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error) {
	return nil, ErrIncompatibleInput
}

func (*nodeServiceContext) OutputData(ctx context.Context, deps *Dependencies, workflow Workflows) (interface{}, error) {
	return nil, nil
}

func (*nodeServiceContext) GetCookies(ctx context.Context, deps *Dependencies, workflow Workflows) ([]*http.Cookie, error) {
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

		ctx := context.Background()
		deps := &Dependencies{}
		logger := ServiceLogger{log.Null}
		database := &db.MockHandle{}
		store := NewMockStore(ctrl)
		uiInfoResolver := NewMockServiceUIInfoResolver(ctrl)

		service := &Service{
			Deps:           deps,
			Logger:         logger,
			Store:          store,
			Database:       database,
			UIInfoResolver: uiInfoResolver,
		}

		Convey("Populate context", func() {
			rng = rand.New(rand.NewSource(0))

			intent := &intentServiceContext{}

			store.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Return(nil)
			store.EXPECT().CreateWorkflow(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			store.EXPECT().DeleteSession(gomock.Any(), gomock.Any()).Return(nil)
			store.EXPECT().DeleteWorkflow(gomock.Any(), gomock.Any()).Return(nil)

			output, err := service.CreateNewWorkflow(ctx, intent, &SessionOptions{
				ClientID: "client-id",
			})
			So(err, ShouldBeNil)

			store.EXPECT().GetSession(gomock.Any(), output.Workflow.WorkflowID).Return(output.Session, nil)
			store.EXPECT().GetWorkflowByInstanceID(gomock.Any(), output.Workflow.InstanceID).Times(2).Return(output.Workflow, nil)

			output, err = service.FeedInput(
				ctx,
				output.Session.WorkflowID,
				output.Workflow.InstanceID,
				"",
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
