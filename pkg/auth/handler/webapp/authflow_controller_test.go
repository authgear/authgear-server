package webapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authflowclient"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func TestAuthflowControllerGetOrCreateWebSession(t *testing.T) {
	Convey("AuthflowController.getOrCreateWebSession", t, func() {
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

			s, err := c.getOrCreateWebSession(w, r, opts)
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
			ss, err := c.getOrCreateWebSession(w, r, opts)
			So(err, ShouldBeNil)
			So(ss, ShouldEqual, s)
		})
	})
}

func TestAuthflowControllerGetScreen(t *testing.T) {
	Convey("AuthflowController.getScreen", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuthflows := NewMockAuthflowControllerAuthflowHTTPClient(ctrl)

		c := &AuthflowController{
			Authflows: mockAuthflows,
		}

		Convey("return ErrFlowNotFound if session has no authflow", func() {
			s := &webapp.Session{}

			_, err := c.getScreen(s, "")
			So(err, ShouldBeError, authflowclient.ErrFlowNotFound)
		})

		Convey("return ErrFlowNotFound if x_step is absent", func() {
			state := &webapp.AuthflowStateToken{
				XStep:      "step_0",
				StateToken: "authflowstate_0",
			}
			screen := &webapp.AuthflowScreen{
				StateToken: state,
			}
			s := &webapp.Session{
				Authflow: &webapp.Authflow{
					AllScreens: map[string]*webapp.AuthflowScreen{
						state.XStep: screen,
					},
				},
			}

			_, err := c.getScreen(s, "")
			So(errors.Is(err, authflowclient.ErrFlowNotFound), ShouldBeTrue)
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
					AllScreens: map[string]*webapp.AuthflowScreen{
						"step_0": screen0,
						"step_1": screen1,
					},
				},
			}

			mockAuthflows.EXPECT().Get("authflowstate_1").Times(1).Return(&authflowclient.FlowResponse{
				StateToken: "authflowstate_1",
				Type:       authflowclient.FlowTypeLogin,
				Name:       "default",
			}, nil)

			actual, err := c.getScreen(s, "step_1")
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, &webapp.AuthflowScreenWithFlowResponse{
				Screen: screen1,
				StateTokenFlowResponse: &authflowclient.FlowResponse{
					StateToken: "authflowstate_1",
					Type:       authflowclient.FlowTypeLogin,
					Name:       "default",
				},
			})
		})
	})
}

func TestAuthflowControllerCreateScreen(t *testing.T) {
	Convey("AuthflowController.createScreen", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuthflows := NewMockAuthflowControllerAuthflowHTTPClient(ctrl)
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
			w := httptest.NewRecorder()

			mockAuthflows.EXPECT().Create(gomock.Any(), gomock.Any()).Times(1).Return(&authflowclient.FlowResponse{
				StateToken: "authflowstate_0",
				Type:       authflowclient.FlowTypeLogin,
				Name:       "default",
				Action: &authflowclient.FlowAction{
					Type: authflowclient.FlowActionTypeIdentify,
				},
			}, nil)
			mockSessionStore.EXPECT().Update(gomock.Any()).Times(1).Return(nil)

			screen, err := c.createScreen(w, r, s, authflowclient.FlowReference{
				Type: authflowclient.FlowTypeLogin,
				Name: "default",
			})
			So(err, ShouldBeNil)
			So(screen, ShouldNotBeNil)
			So(string(screen.BranchStateTokenFlowResponse.Type), ShouldEqual, "login")
		})
	})
}

func TestAuthflowControllerFeedInput(t *testing.T) {
	Convey("AuthflowController.FeedInput", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuthflows := NewMockAuthflowControllerAuthflowHTTPClient(ctrl)
		mockSessionStore := NewMockAuthflowControllerSessionStore(ctrl)
		mockClock := clock.NewMockClockAt("2006-01-02T03:04:05Z")

		c := AuthflowController{
			Clock:     mockClock,
			Authflows: mockAuthflows,
			Sessions:  mockSessionStore,
		}

		Convey("the branch does not require input to take", func() {
			r, _ := http.NewRequest("POST", "", nil)
			w := httptest.NewRecorder()
			s := &webapp.Session{
				Authflow: &webapp.Authflow{
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
				StateTokenFlowResponse: &authflowclient.FlowResponse{
					StateToken: "authflowstate_0",
					Type:       authflowclient.FlowTypeLogin,
					Name:       "default",
					Action: &authflowclient.FlowAction{

						Type: authflowclient.FlowActionTypeIdentify,
					},
				},
			}

			input := map[string]interface{}{
				"identification": "email",
				"login_id":       "johndoe@example.com",
			}

			mockAuthflows.EXPECT().Input(w, r, "authflowstate_0", gomock.Any()).Times(1).Return(&authflowclient.FlowResponse{
				StateToken: "authflowstate_1",
				Type:       authflowclient.FlowTypeLogin,
				Name:       "default",
				Action: &authflowclient.FlowAction{
					Type: authflowclient.FlowActionTypeAuthenticate,
					Data: json.RawMessage(`{"options":[{"authentication":"primary_password"}]}`),
				},
			}, nil)
			mockSessionStore.EXPECT().Update(s).Times(1).Return(nil)

			result, err := c.AdvanceWithInput(w, r, s, screen, input)
			So(err, ShouldBeNil)
			So(strings.HasPrefix(result.RedirectURI, "/authflow/enter_password?x_step="), ShouldBeTrue)
		})

		Convey("the branch requires input to take", func() {
			r, _ := http.NewRequest("POST", "", nil)
			w := httptest.NewRecorder()
			s := &webapp.Session{
				Authflow: &webapp.Authflow{
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
				StateTokenFlowResponse: &authflowclient.FlowResponse{
					StateToken: "authflowstate_0",
					Type:       authflowclient.FlowTypeLogin,
					Name:       "default",
					Action: &authflowclient.FlowAction{

						Type: authflowclient.FlowActionTypeIdentify,
					},
				},
			}

			input := map[string]interface{}{
				"identification": "email",
				"login_id":       "johndoe@example.com",
			}

			gomock.InOrder(
				mockAuthflows.EXPECT().Input(w, r, "authflowstate_0", input).Times(1).Return(&authflowclient.FlowResponse{
					StateToken: "authflowstate_1",
					Type:       authflowclient.FlowTypeLogin,
					Name:       "default",
					Action: &authflowclient.FlowAction{
						Type: authflowclient.FlowActionTypeAuthenticate,
						Data: json.RawMessage(`{"options":[{"authentication":"primary_oob_otp_email","channels":["email"]}]}`),
					},
				}, nil),
				mockAuthflows.EXPECT().Input(w, r, "authflowstate_1", jsonMatcher{
					v: map[string]interface{}{
						"authentication": "primary_oob_otp_email",
						"index":          0,
						"channel":        "email",
					},
				}).Times(1).Return(&authflowclient.FlowResponse{
					StateToken: "authflowstate_2",
					Type:       authflowclient.FlowTypeLogin,
					Name:       "default",
					Action: &authflowclient.FlowAction{
						Type:           authflowclient.FlowActionTypeAuthenticate,
						Authentication: authflowclient.AuthenticationPrimaryOOBOTPEmail,
						Data:           json.RawMessage(`{"otp_form":"code"}`),
					},
				}, nil),
				mockSessionStore.EXPECT().Update(s).Times(1).Return(nil),
			)

			result, err := c.AdvanceWithInput(w, r, s, screen, input)
			So(err, ShouldBeNil)
			So(strings.HasPrefix(result.RedirectURI, "/authflow/enter_oob_otp?x_step="), ShouldBeTrue)
		})
	})
}

type jsonMatcher struct {
	v interface{}
}

func (m jsonMatcher) Matches(x interface{}) bool {
	aJSONStr, err := json.Marshal(m.v)
	if err != nil {
		return false
	}
	bJSONStr, err := json.Marshal(x)
	if err != nil {
		return false
	}

	return string(aJSONStr) == string(bJSONStr)
}

func (m jsonMatcher) String() string {
	return fmt.Sprintf("%v", m.v)
}
