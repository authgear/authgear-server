package authenticationflow

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
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
		uiInfoResolver := NewMockServiceUIInfoResolver(ctrl)

		service := &Service{
			ContextDoNotUseDirectly: ctx,
			Deps:                    deps,
			Logger:                  logger,
			Store:                   store,
			Database:                database,
			UIInfoResolver:          uiInfoResolver,
		}

		Convey("CreateNewFlow with intent expecting non-nil input at the beginning", func() {
			rng = rand.New(rand.NewSource(0))

			intent := &intentAuthenticate{
				PretendLoginIDExists: false,
			}

			gomock.InOrder(
				store.EXPECT().CreateSession(gomock.Any()).Return(nil),
				store.EXPECT().CreateFlow(gomock.Any()).Return(nil),
			)

			output, err := service.CreateNewFlow(intent, &SessionOptions{})
			So(err, ShouldBeNil)
			schemaBuilder := validation.SchemaBuilder{}.
				Type(validation.TypeObject).
				Required("login_id")

			schemaBuilder.Properties().Property(
				"login_id",
				validation.SchemaBuilder{}.Type(validation.TypeString),
			)

			So(output, ShouldResemble, &ServiceOutput{
				Flow: &Flow{
					FlowID:     "authflowparent_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					InstanceID: "authflow_1WPH8EXJFWMAZ7M8Y9EGAG34SPW86VXT",
					Intent:     intent,
				},
				Data:          &DataFlowDetails{},
				SchemaBuilder: schemaBuilder,
				Session: &Session{
					FlowID: "authflowparent_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
				},
				SessionOutput: &SessionOutput{
					FlowID: "authflowparent_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
				},
			})
		})

		SkipConvey("CreateNewFlow with intent expecting nil input at the beginning", func() {
			rng = rand.New(rand.NewSource(0))

			intent := &intentNilInput{}

			gomock.InOrder(
				store.EXPECT().CreateSession(gomock.Any()).Return(nil),
				store.EXPECT().CreateFlow(gomock.Any()).Return(nil),

				uiInfoResolver.EXPECT().SetAuthenticationInfoInQuery(gomock.Any(), gomock.Any()).Return(""),

				store.EXPECT().DeleteSession(gomock.Any()).Return(nil),
				store.EXPECT().DeleteFlow(gomock.Any()).Return(nil),
			)

			output, err := service.CreateNewFlow(intent, &SessionOptions{})
			So(errors.Is(err, ErrEOF), ShouldBeTrue)
			So(output, ShouldResemble, &ServiceOutput{
				Flow: &Flow{
					FlowID:     "authflowparent_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					InstanceID: "authflow_Y37GSHFPM7259WFBY64B4HTJ4PM8G482",
					Intent:     intent,
					Nodes: []Node{
						{
							Type:   NodeTypeSimple,
							Simple: &nodeNilInput{},
						},
					},
				},
				Finished: true,
				Data: mapData{
					"flow_reference": map[string]interface{}{
						"type": "",
						"id":   "",
					},
				},
				Session: &Session{
					FlowID: "authflowparent_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
				},
				SessionOutput: &SessionOutput{
					FlowID: "authflowparent_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
				},
			})
		})

		SkipConvey("FeedInput", func() {
			rng = rand.New(rand.NewSource(0))

			intent := &intentAuthenticate{
				PretendLoginIDExists: false,
			}
			flow := &Flow{
				FlowID:     "flow-id",
				InstanceID: "instance-id",
				Intent:     intent,
			}
			session := &Session{
				FlowID: "flow-id",
			}

			gomock.InOrder(
				store.EXPECT().GetFlowByInstanceID(flow.InstanceID).Times(1).Return(flow, nil),
				store.EXPECT().GetSession(flow.FlowID).Return(session, nil),
				store.EXPECT().GetFlowByInstanceID(flow.InstanceID).Times(1).Return(flow, nil),
				store.EXPECT().CreateFlow(gomock.Any()).Return(nil),
			)

			output, err := service.FeedInput(flow.InstanceID, "", json.RawMessage(`{
				"login_id": "user@example.com"
			}`))
			So(err, ShouldBeNil)
			So(output, ShouldResemble, &ServiceOutput{
				Flow: &Flow{
					FlowID:     "flow-id",
					InstanceID: "authflow_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					Intent:     intent,
					Nodes: []Node{
						{
							Type: NodeTypeSubFlow,
							SubFlow: &Flow{
								Intent: &intentSignup{
									LoginID: "user@example.com",
								},
								Nodes: []Node{
									{
										Type: NodeTypeSubFlow,

										SubFlow: &Flow{
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
				Data: &DataFlowDetails{},
				Session: &Session{
					FlowID: "flow-id",
				},
				SessionOutput: &SessionOutput{
					FlowID: "flow-id",
				},
			})
		})
	})
}

type intentNilInput struct{}

var _ PublicFlow = &intentNilInput{}

func (*intentNilInput) Kind() string {
	return "intentNilInput"
}

func (*intentNilInput) FlowType() FlowType {
	return ""
}

func (*intentNilInput) FlowInit(r FlowReference) {}

func (*intentNilInput) FlowFlowReference() FlowReference {
	return FlowReference{}
}

func (*intentNilInput) FlowRootObject(deps *Dependencies) (config.AuthenticationFlowObject, error) {
	return nil, nil
}

func (*intentNilInput) CanReactTo(ctx context.Context, deps *Dependencies, flows Flows) (InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 {
		return nil, nil
	}
	return nil, ErrEOF
}

func (*intentNilInput) ReactTo(ctx context.Context, deps *Dependencies, flows Flows, input Input) (*Node, error) {
	return NewNodeSimple(&nodeNilInput{}), nil
}

type nodeNilInput struct {
	ClientID string
}

func (*nodeNilInput) Kind() string {
	return "nodeNilInput"
}

type intentServiceContext struct{}

var _ PublicFlow = &intentServiceContext{}
var _ CookieGetter = &intentServiceContext{}

func (*intentServiceContext) Kind() string {
	return "intentServiceContext"
}

func (*intentServiceContext) FlowType() FlowType {
	return ""
}

func (*intentServiceContext) FlowInit(r FlowReference) {}

func (*intentServiceContext) FlowFlowReference() FlowReference {
	return FlowReference{}
}

func (*intentServiceContext) FlowRootObject(deps *Dependencies) (config.AuthenticationFlowObject, error) {
	return nil, nil
}

func (*intentServiceContext) CanReactTo(ctx context.Context, deps *Dependencies, flows Flows) (InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 {
		return &inputServiceContext{}, nil
	}
	return nil, ErrEOF
}

func (*intentServiceContext) ReactTo(ctx context.Context, deps *Dependencies, flows Flows, input Input) (*Node, error) {
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

func (*intentServiceContext) GetCookies(ctx context.Context, deps *Dependencies, flows Flows) ([]*http.Cookie, error) {
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

func (*inputServiceContext) GetJSONPointer() jsonpointer.T {
	return nil
}

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

func (*nodeServiceContext) GetCookies(ctx context.Context, deps *Dependencies, flows Flows) ([]*http.Cookie, error) {
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
		uiInfoResolver := NewMockServiceUIInfoResolver(ctrl)

		service := &Service{
			ContextDoNotUseDirectly: ctx,
			Deps:                    deps,
			Logger:                  logger,
			Store:                   store,
			Database:                database,
			UIInfoResolver:          uiInfoResolver,
		}

		Convey("Populate context", func() {
			rng = rand.New(rand.NewSource(0))

			intent := &intentServiceContext{}

			store.EXPECT().CreateSession(gomock.Any()).Return(nil)
			store.EXPECT().CreateFlow(gomock.Any()).AnyTimes().Return(nil)
			store.EXPECT().DeleteSession(gomock.Any()).Return(nil)
			store.EXPECT().DeleteFlow(gomock.Any()).Return(nil)

			output, err := service.CreateNewFlow(intent, &SessionOptions{
				ClientID: "client-id",
			})
			So(err, ShouldBeNil)

			store.EXPECT().GetSession(output.Flow.FlowID).Return(output.Session, nil)
			store.EXPECT().GetFlowByInstanceID(output.Flow.InstanceID).Times(2).Return(output.Flow, nil)

			output, err = service.FeedInput(
				output.Flow.InstanceID,
				"",
				json.RawMessage(`{}`),
			)
			So(errors.Is(err, ErrEOF), ShouldBeTrue)
			So(output, ShouldResemble, &ServiceOutput{
				Flow: &Flow{
					FlowID:     "authflowparent_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					InstanceID: "authflow_Y37GSHFPM7259WFBY64B4HTJ4PM8G482",
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
				Finished: true,
				Data: mapData{
					"flow_reference": map[string]interface{}{
						"type": "",
						"id":   "",
					},
				},
				Session: &Session{
					FlowID:   "authflowparent_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					ClientID: "client-id",
				},
				SessionOutput: &SessionOutput{
					FlowID:   "authflowparent_TJSAV0F58G8VBWREZ22YBMAW1A0GFCD4",
					ClientID: "client-id",
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
