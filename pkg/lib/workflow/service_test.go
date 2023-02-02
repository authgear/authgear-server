package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/log"
)

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

		Convey("CreateNewWorkflow", func() {
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

type intentServiceContext struct{}

func (*intentServiceContext) Kind() string {
	return "intentServiceContext"
}

func (i *intentServiceContext) Instantiate(data json.RawMessage) error {
	return json.Unmarshal(data, i)
}

func (*intentServiceContext) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (*intentServiceContext) DeriveEdges(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Edge, error) {
	if len(workflow.Nodes) == 0 {
		return []Edge{
			&edgeServiceContext{},
		}, nil
	}
	return nil, ErrEOF
}

func (*intentServiceContext) OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error) {
	return nil, nil
}

type edgeServiceContext struct{}

func (*edgeServiceContext) Instantiate(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
	return NewNodeSimple(&nodeServiceContext{
		ClientID: GetClientID(ctx),
	}), nil
}

type nodeServiceContext struct {
	ClientID string
}

func (*nodeServiceContext) Kind() string {
	return "nodeServiceContext"
}

func (*nodeServiceContext) GetEffects(ctx context.Context, deps *Dependencies) ([]Effect, error) {
	return nil, nil
}

func (*nodeServiceContext) DeriveEdges(ctx context.Context, deps *Dependencies) ([]Edge, error) {
	return nil, ErrEOF
}

func (*nodeServiceContext) OutputData(ctx context.Context, deps *Dependencies) (interface{}, error) {
	return nil, nil
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
			store.EXPECT().GetWorkflowByInstanceID(output.Workflow.InstanceID).Times(1).Return(output.Workflow, nil)

			output, err = service.FeedInput(
				output.Workflow.WorkflowID,
				output.Workflow.InstanceID,
				nil,
			)
			So(errors.Is(err, ErrEOF), ShouldBeTrue)
			So(output, ShouldResemble, &ServiceOutput{
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
			})
		})
	})
}
