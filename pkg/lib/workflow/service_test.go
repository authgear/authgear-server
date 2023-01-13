package workflow

import (
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

		context := &Context{}
		logger := ServiceLogger{log.Null}
		savepoint := NewMockSavepoint(ctrl)
		store := NewMockStore(ctrl)

		service := &Service{
			Context:   context,
			Logger:    logger,
			Store:     store,
			Savepoint: savepoint,
		}

		Convey("CreateNewWorkflow", func() {
			rng = rand.New(rand.NewSource(0))

			intent := &intentAuthenticate{
				PretendLoginIDExists: false,
			}

			gomock.InOrder(
				savepoint.EXPECT().Begin().Times(1).Return(nil),
				store.EXPECT().CreateWorkflow(gomock.Any()).Return(nil),
				savepoint.EXPECT().Rollback().Times(1).Return(nil),
				store.EXPECT().CreateSession(gomock.Any()).Return(nil),
			)

			output, err := service.CreateNewWorkflow(intent)
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
				savepoint.EXPECT().Begin().Times(1).Return(nil),
				store.EXPECT().GetWorkflowByInstanceID(workflow.InstanceID).Times(1).Return(workflow, nil),
				store.EXPECT().CreateWorkflow(gomock.Any()).Return(nil),
				savepoint.EXPECT().Rollback().Times(1).Return(nil),
				store.EXPECT().GetSession(workflow.WorkflowID).Return(session, nil),
			)

			output, err := service.FeedInput(workflow.InstanceID, &inputLoginID{
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
