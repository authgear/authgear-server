package workflow2

import (
	"context"
	"encoding/json"
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
	RegisterIntent(&intentServiceContext{})
	RegisterIntent(&intentNilInput{})
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
		database := &db.MockHandle{}
		store := NewMockStore(ctrl)

		service := &Service{
			ContextDoNotUseDirectly: ctx,
			Deps:                    deps,
			Logger:                  logger,
			Store:                   store,
			Database:                database,
		}

		Convey("CreateNewWorkflow with intent expecting non-nil input at the beginning", func() {
			rng = rand.New(rand.NewSource(0))

			intent := &intentAuthenticate{
				PretendLoginIDExists: false,
			}

			gomock.InOrder(
				store.EXPECT().CreateSession(gomock.Any()).Return(nil),
				store.EXPECT().CreateWorkflow(gomock.Any()).Return(nil),
			)

			output, err := service.CreateNewWorkflow(intent, &SessionOptions{})
			So(err, ShouldBeNil)
			schemaBuilder := validation.SchemaBuilder{}.
				Type(validation.TypeObject).
				Required("login_id")

			schemaBuilder.Properties().Property(
				"login_id",
				validation.SchemaBuilder{}.Type(validation.TypeString),
			)

			So(output, ShouldResemble, &ServiceOutput{
				Workflow: &Workflow{
					WorkflowID: "flowparent_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					InstanceID: "flow_1WPH8EXJFWMAZ7M8Y9EGAG34SPW86VXT",
					Intent:     intent,
				},
				Data:          EmptyData,
				SchemaBuilder: schemaBuilder,
				Session: &Session{
					WorkflowID: "flowparent_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
				},
				SessionOutput: &SessionOutput{
					WorkflowID: "flowparent_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
				},
			})
		})

		SkipConvey("CreateNewWorkflow with intent expecting nil input at the beginning", func() {
			rng = rand.New(rand.NewSource(0))

			intent := &intentNilInput{}

			gomock.InOrder(
				store.EXPECT().CreateSession(gomock.Any()).Return(nil),
				store.EXPECT().CreateWorkflow(gomock.Any()).Return(nil),

				store.EXPECT().DeleteSession(gomock.Any()).Return(nil),
				store.EXPECT().DeleteWorkflow(gomock.Any()).Return(nil),
			)

			output, err := service.CreateNewWorkflow(intent, &SessionOptions{})
			So(errors.Is(err, ErrEOF), ShouldBeTrue)
			So(output, ShouldResemble, &ServiceOutput{
				Workflow: &Workflow{
					WorkflowID: "flowparent_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					InstanceID: "flow_Y37GSHFPM7259WFBY64B4HTJ4PM8G482",
					Intent:     intent,
					Nodes: []Node{
						{
							Type:   NodeTypeSimple,
							Simple: &nodeNilInput{},
						},
					},
				},
				Data: &DataRedirectURI{
					RedirectURI: "",
				},
				Session: &Session{
					WorkflowID: "flowparent_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
				},
				SessionOutput: &SessionOutput{
					WorkflowID: "flowparent_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
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
				store.EXPECT().GetWorkflowByInstanceID(workflow.InstanceID).Times(1).Return(workflow, nil),
				store.EXPECT().GetSession(workflow.WorkflowID).Return(session, nil),
				store.EXPECT().GetWorkflowByInstanceID(workflow.InstanceID).Times(1).Return(workflow, nil),
				store.EXPECT().CreateWorkflow(gomock.Any()).Return(nil),
			)

			output, err := service.FeedInput(workflow.InstanceID, "", json.RawMessage(`{
				"login_id": "user@example.com"
			}`))
			So(err, ShouldBeNil)
			So(output, ShouldResemble, &ServiceOutput{
				Workflow: &Workflow{
					WorkflowID: "workflow-id",
					InstanceID: "flow_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
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
				Data: EmptyData,
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

var _ Intent = &intentNilInput{}

func (*intentNilInput) Kind() string {
	return "intentNilInput"
}

func (*intentNilInput) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) (InputSchema, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return nil, nil
	}
	return nil, ErrEOF
}

func (*intentNilInput) ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error) {
	return NewNodeSimple(&nodeNilInput{}), nil
}

type nodeNilInput struct {
	ClientID string
}

func (*nodeNilInput) Kind() string {
	return "nodeNilInput"
}

type intentServiceContext struct{}

var _ Intent = &intentServiceContext{}
var _ CookieGetter = &intentServiceContext{}

func (*intentServiceContext) Kind() string {
	return "intentServiceContext"
}

func (*intentServiceContext) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) (InputSchema, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return &inputServiceContext{}, nil
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

func (*intentServiceContext) GetCookies(ctx context.Context, deps *Dependencies, workflows Workflows) ([]*http.Cookie, error) {
	return []*http.Cookie{
		{
			Name:  "intentServiceContext",
			Value: "intentServiceContext",
		},
	}, nil
}

type inputServiceContext struct{}

var _ InputSchema = &inputServiceContext{}
var _ Input = &inputServiceContext{}
var _ InputServiceContext = &inputServiceContext{}

func (*inputServiceContext) SchemaBuilder() validation.SchemaBuilder {
	return validation.SchemaBuilder{}
}

func (*inputServiceContext) MakeInput(rawMessage json.RawMessage) (Input, error) {
	var input inputServiceContext
	err := json.Unmarshal(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func (*inputServiceContext) Input() {}

func (*inputServiceContext) Marker() {}

type InputServiceContext interface {
	Marker()
}

type nodeServiceContext struct {
	ClientID string
}

var _ NodeSimple = &nodeServiceContext{}
var _ CookieGetter = &nodeServiceContext{}

func (*nodeServiceContext) Kind() string {
	return "nodeServiceContext"
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

		ctx := context.TODO()
		deps := &Dependencies{}
		logger := ServiceLogger{log.Null}
		database := &db.MockHandle{}
		store := NewMockStore(ctrl)

		service := &Service{
			ContextDoNotUseDirectly: ctx,
			Deps:                    deps,
			Logger:                  logger,
			Store:                   store,
			Database:                database,
		}

		Convey("Populate context", func() {
			rng = rand.New(rand.NewSource(0))

			intent := &intentServiceContext{}

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
				output.Workflow.InstanceID,
				"",
				json.RawMessage(`{}`),
			)
			So(errors.Is(err, ErrEOF), ShouldBeTrue)
			So(output, ShouldResemble, &ServiceOutput{
				Workflow: &Workflow{
					WorkflowID: "flowparent_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					InstanceID: "flow_Y37GSHFPM7259WFBY64B4HTJ4PM8G482",
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
				Data: &DataRedirectURI{
					RedirectURI: "",
				},
				Session: &Session{
					WorkflowID: "flowparent_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					ClientID:   "client-id",
				},
				SessionOutput: &SessionOutput{
					WorkflowID: "flowparent_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
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
