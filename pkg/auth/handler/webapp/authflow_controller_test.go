package webapp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func TestAuthflowControllerGetOrCreateWebSession(t *testing.T) {
	Convey("AuthflowController.GetOrCreateWebSession", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSessionStore := NewMockAuthflowControllerSessionStore(ctrl)
		mockCookieManager := NewMockAuthflowControllerCookieManager(ctrl)

		c := &AuthflowController{
			Clock:         clock.NewMockClockAt("2006-01-02T03:04:05Z"),
			Cookies:       mockCookieManager,
			Sessions:      mockSessionStore,
			SessionCookie: webapp.NewSessionCookieDef(),
		}

		Convey("Create new if not in context", func() {
			ctx := context.Background()
			r, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

			w := httptest.NewRecorder()

			opts := webapp.SessionOptions{}

			mockSessionStore.EXPECT().Create(gomock.Any()).Times(1).Return(nil)
			mockCookieManager.EXPECT().ValueCookie(c.SessionCookie.Def, gomock.Any()).Times(1).Return(&http.Cookie{})

			s, err := c.GetOrCreateWebSession(w, r, opts)
			So(err, ShouldBeNil)
			So(s, ShouldNotBeNil)
		})

		Convey("return session in context", func() {
			ctx := context.Background()
			s := &webapp.Session{
				ID: "test",
			}
			ctx = webapp.WithSession(ctx, s)

			r, _ := http.NewRequestWithContext(ctx, "GET", "", nil)

			w := httptest.NewRecorder()

			opts := webapp.SessionOptions{}
			ss, err := c.GetOrCreateWebSession(w, r, opts)
			So(err, ShouldBeNil)
			So(ss, ShouldEqual, s)
		})
	})
}

func TestAuthflowControllerGetScreen(t *testing.T) {
	Convey("AuthflowController.GetScreen", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuthflows := NewMockAuthflowControllerAuthflowService(ctrl)

		c := &AuthflowController{
			Authflows: mockAuthflows,
		}

		Convey("return ErrFlowNotFound if session has no authflow", func() {
			s := &webapp.Session{}

			_, err := c.GetScreen(s, "")
			So(err, ShouldBeError, authflow.ErrFlowNotFound)
		})

		Convey("return ErrFlowNotFound if x_step is absent", func() {
			serviceOutput := &authflow.ServiceOutput{
				Flow: &authflow.Flow{
					FlowID:     "authflow_id",
					StateToken: "authflowstate_0",
				},
				FlowReference: &authflow.FlowReference{
					Type: authflow.FlowTypeLogin,
					Name: "default",
				},
			}
			flowResponse := serviceOutput.ToFlowResponse()
			state := &webapp.AuthflowStateToken{
				XStep:      "step_0",
				StateToken: flowResponse.StateToken,
			}
			screen := &webapp.AuthflowScreen{
				StateToken: state,
			}
			s := &webapp.Session{
				Authflow: &webapp.Authflow{
					FlowID: flowResponse.ID,
					AllScreens: map[string]*webapp.AuthflowScreen{
						state.XStep: screen,
					},
				},
			}

			_, err := c.GetScreen(s, "")
			So(errors.Is(err, authflow.ErrFlowNotFound), ShouldBeTrue)
		})

		Convey("return screen as specified by x_step", func() {
			screen0 := &webapp.AuthflowScreen{
				StateToken: &webapp.AuthflowStateToken{
					XStep:      "step_0",
					StateToken: "authflowstate_0",
				},
			}
			screen1 := &webapp.AuthflowScreen{
				StateToken: &webapp.AuthflowStateToken{
					XStep:      "step_1",
					StateToken: "authflowstate_1",
				},
			}

			s := &webapp.Session{
				Authflow: &webapp.Authflow{
					FlowID: "authflow_id",
					AllScreens: map[string]*webapp.AuthflowScreen{
						"step_0": screen0,
						"step_1": screen1,
					},
				},
			}

			mockAuthflows.EXPECT().Get("authflowstate_1").Times(1).Return(&authflow.ServiceOutput{
				Flow: &authflow.Flow{
					FlowID:     "authflow_id",
					StateToken: "authflowstate_1",
				},
				FlowReference: &authflow.FlowReference{
					Type: authflow.FlowTypeLogin,
					Name: "default",
				},
			}, nil)

			actual, err := c.GetScreen(s, "step_1")
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, &webapp.AuthflowScreenWithFlowResponse{
				Screen: screen1,
				StateTokenFlowResponse: &authflow.FlowResponse{
					ID:         "authflow_id",
					StateToken: "authflowstate_1",
					Type:       authflow.FlowTypeLogin,
					Name:       "default",
				},
			})
		})
	})
}

func TestAuthflowControllerCreateScreen(t *testing.T) {
	Convey("AuthflowController.CreateScreen", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuthflows := NewMockAuthflowControllerAuthflowService(ctrl)
		mockSessionStore := NewMockAuthflowControllerSessionStore(ctrl)
		mockClock := clock.NewMockClockAt("2006-01-02T03:04:05Z")

		c := &AuthflowController{
			Clock:     mockClock,
			Authflows: mockAuthflows,
			Sessions:  mockSessionStore,
		}

		Convey("create screen", func() {
			ctx := context.Background()
			s := &webapp.Session{}

			r, _ := http.NewRequestWithContext(ctx, "GET", "/authflow/login", nil)

			mockAuthflows.EXPECT().CreateNewFlow(gomock.Any(), gomock.Any()).Times(1).Return(&authflow.ServiceOutput{
				Flow: &authflow.Flow{
					FlowID:     "authflow_id",
					StateToken: "authflowstate_0",
				},
				FlowReference: &authflow.FlowReference{
					Type: authflow.FlowTypeLogin,
					Name: "default",
				},
				FlowAction: &authflow.FlowAction{
					Type: authflow.FlowActionTypeFromStepType(config.AuthenticationFlowStepTypeIdentify),
				},
			}, nil)
			mockSessionStore.EXPECT().Update(gomock.Any()).Times(1).Return(nil)

			result, err := c.CreateScreen(r, s, authflow.FlowReference{
				Type: authflow.FlowTypeLogin,
				Name: "default",
			})
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.NavigationAction, ShouldEqual, "replace")
			So(strings.HasPrefix(result.RedirectURI, "/authflow/login?x_step="), ShouldBeTrue)
		})
	})
}

func TestAuthflowControllerFeedInput(t *testing.T) {
	Convey("AuthflowController.FeedInput", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuthflows := NewMockAuthflowControllerAuthflowService(ctrl)
		mockSessionStore := NewMockAuthflowControllerSessionStore(ctrl)
		mockClock := clock.NewMockClockAt("2006-01-02T03:04:05Z")

		c := AuthflowController{
			Clock:     mockClock,
			Authflows: mockAuthflows,
			Sessions:  mockSessionStore,
		}

		Convey("the branch does not require input to take", func() {
			r, _ := http.NewRequest("POST", "", nil)
			s := &webapp.Session{
				Authflow: &webapp.Authflow{
					FlowID:     "authflow_id",
					AllScreens: map[string]*webapp.AuthflowScreen{},
				},
			}

			screen := &webapp.AuthflowScreenWithFlowResponse{
				Screen: &webapp.AuthflowScreen{
					StateToken: &webapp.AuthflowStateToken{
						XStep:      "step_0",
						StateToken: "authflowstate_0",
					},
					BranchStateToken: &webapp.AuthflowStateToken{
						XStep:      "step_0",
						StateToken: "authflowstate_0",
					},
				},
				StateTokenFlowResponse: &authflow.FlowResponse{
					ID:         "authflow_id",
					StateToken: "authflowstate_0",
					Type:       authflow.FlowTypeLogin,
					Name:       "default",
					Action: &authflow.FlowAction{

						Type: authflow.FlowActionTypeFromStepType(config.AuthenticationFlowStepTypeIdentify),
					},
				},
			}

			input := map[string]interface{}{
				"identification": "email",
				"login_id":       "johndoe@example.com",
			}

			mockAuthflows.EXPECT().FeedInput("authflowstate_0", gomock.Any()).Times(1).Return(&authflow.ServiceOutput{
				Flow: &authflow.Flow{
					FlowID:     "authflow_id",
					StateToken: "authflowstate_1",
				},
				FlowReference: &authflow.FlowReference{
					Type: authflow.FlowTypeLogin,
					Name: "default",
				},
				FlowAction: &authflow.FlowAction{
					Type: authflow.FlowActionTypeFromStepType(config.AuthenticationFlowStepTypeAuthenticate),
					Data: declarative.IntentLoginFlowStepAuthenticateData{
						Options: []declarative.UseAuthenticationOption{
							{
								Authentication: config.AuthenticationFlowAuthenticationPrimaryPassword,
							},
						},
					},
				},
			}, nil)
			mockSessionStore.EXPECT().Update(s).Times(1).Return(nil)

			result, err := c.FeedInput(r, s, screen, input)
			So(err, ShouldBeNil)
			So(strings.HasPrefix(result.RedirectURI, "/authflow/enter_password?x_step="), ShouldBeTrue)
		})

		Convey("the branch requires input to take", func() {
			r, _ := http.NewRequest("POST", "", nil)
			s := &webapp.Session{
				Authflow: &webapp.Authflow{
					FlowID:     "authflow_id",
					AllScreens: map[string]*webapp.AuthflowScreen{},
				},
			}

			screen := &webapp.AuthflowScreenWithFlowResponse{
				Screen: &webapp.AuthflowScreen{
					StateToken: &webapp.AuthflowStateToken{
						XStep:      "step_0",
						StateToken: "authflowstate_0",
					},
					BranchStateToken: &webapp.AuthflowStateToken{
						XStep:      "step_0",
						StateToken: "authflowstate_0",
					},
				},
				StateTokenFlowResponse: &authflow.FlowResponse{
					ID:         "authflow_id",
					StateToken: "authflowstate_0",
					Type:       authflow.FlowTypeLogin,
					Name:       "default",
					Action: &authflow.FlowAction{

						Type: authflow.FlowActionTypeFromStepType(config.AuthenticationFlowStepTypeIdentify),
					},
				},
			}

			input := map[string]interface{}{
				"identification": "email",
				"login_id":       "johndoe@example.com",
			}

			mockAuthflows.EXPECT().FeedInput("authflowstate_0", gomock.Any()).Times(1).Return(&authflow.ServiceOutput{
				Flow: &authflow.Flow{
					FlowID:     "authflow_id",
					StateToken: "authflowstate_1",
				},
				FlowReference: &authflow.FlowReference{
					Type: authflow.FlowTypeLogin,
					Name: "default",
				},
				FlowAction: &authflow.FlowAction{
					Type: authflow.FlowActionTypeFromStepType(config.AuthenticationFlowStepTypeAuthenticate),
					Data: declarative.IntentLoginFlowStepAuthenticateData{
						Options: []declarative.UseAuthenticationOption{
							{
								Authentication: config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
								Channels: []model.AuthenticatorOOBChannel{
									model.AuthenticatorOOBChannelEmail,
								},
							},
						},
					},
				},
			}, nil)
			mockSessionStore.EXPECT().Update(s).Times(1).Return(nil)
			mockAuthflows.EXPECT().FeedInput("authflowstate_1", json.RawMessage(`{"authentication":"primary_oob_otp_email","channel":"email","index":0}`)).Times(1).Return(&authflow.ServiceOutput{
				Flow: &authflow.Flow{
					FlowID:     "authflow_id",
					StateToken: "authflowstate_2",
				},
				FlowReference: &authflow.FlowReference{
					Type: authflow.FlowTypeLogin,
					Name: "default",
				},
				FlowAction: &authflow.FlowAction{
					Type:           authflow.FlowActionTypeFromStepType(config.AuthenticationFlowStepTypeAuthenticate),
					Authentication: config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
					Data:           declarative.NodeVerifyClaimData{},
				},
			}, nil)

			result, err := c.FeedInput(r, s, screen, input)
			So(err, ShouldBeNil)
			So(strings.HasPrefix(result.RedirectURI, "/authflow/enter_oob_otp?x_step="), ShouldBeTrue)
		})
	})
}
