package handler

import (
	"context"
	"time"
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

// AttachSetDisableHandler attaches SetDisableHandler to server
func AttachSetDisableHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/disable/set", &SetDisableHandlerFactory{
		authDependency,
	}).Methods("POST")
	return server
}

// SetDisableHandlerFactory creates SetDisableHandler
type SetDisableHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new SetDisableHandler
func (f SetDisableHandlerFactory) NewHandler(ctx context.Context, tenantConfig config.TenantConfiguration) handler.Handler {
	h := &SetDisableHandler{}
	inject.DefaultInject(h, f.Dependency, ctx, tenantConfig)
	return handler.APIHandlerToHandler(h)
}

type setDisableUserPayload struct {
	AuthInfoID   string `json:"auth_id"`
	Disabled     bool   `json:"disabled"`
	Message      string `json:"message"`
	ExpiryString string `json:"expiry"`
	expiry       *time.Time
}

func (payload setDisableUserPayload) Validate() error {
	if payload.AuthInfoID == "" {
		return skyerr.NewInvalidArgument("invalid auth id", []string{"auth_id"})
	}
	return nil
}

// SetDisableHandler handles set disable request
type SetDisableHandler struct{
	AuthInfoStore        authinfo.Store             `dependency:"AuthInfoStore"`
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h SetDisableHandler) ProvideAuthzPolicy() authz.Policy {
	// FIXME: Admin only after adding admin role
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

// DecodeRequest decode request payload
func (h SetDisableHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := setDisableUserPayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}

	if payload.ExpiryString != "" {
		if expiry, err := time.Parse(time.RFC3339, payload.ExpiryString); err == nil {
			payload.expiry = &expiry
		} else {
			return nil, skyerr.NewInvalidArgument("invalid expiry", []string{"expiry"})
		}
	}

	return payload, nil
}

// Handle function handle set disabled request
func (h SetDisableHandler) Handle(req interface{}, ctx handler.AuthContext) (resp interface{}, err error) {
	p := req.(setDisableUserPayload)

	authinfo := authinfo.AuthInfo{}
	if e := h.AuthInfoStore.GetAuth(p.AuthInfoID, &authinfo); e != nil {
		if err == skydb.ErrUserNotFound {
			// logger.Info("Auth info not found when setting disabled user status")
			err = skyerr.NewError(skyerr.ResourceNotFound, "User not found")
			return
		}
		// logger.WithError(err).Error("Unable to get auth info when setting disabled user status")
		err = skyerr.NewError(skyerr.ResourceNotFound, "User not found")
		return
	}

	authinfo.Disabled = p.Disabled
	if !authinfo.Disabled {
		authinfo.DisabledMessage = ""
		authinfo.DisabledExpiry = nil
	} else {
		authinfo.DisabledMessage = p.Message
		authinfo.DisabledExpiry = p.expiry
	}

	// logger.WithFields(logrus.Fields{
	// 	"disabled": authinfo.Disabled,
	// 	"message":  authinfo.DisabledMessage,
	// 	"expiry":   authinfo.DisabledExpiry,
	// }).Debug("Will set disabled user status")

	if e := h.AuthInfoStore.UpdateAuth(&authinfo); e != nil {
		// logger.WithError(err).Error("Unable to update auth info when setting disabled user status")
		err = skyerr.MakeError(err)
		return
	}

	// logger.Info("Successfully set disabled user status")

	// h.logAuditTrail(payload, p)

	resp = "OK"

	return
}
