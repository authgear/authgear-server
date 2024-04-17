package authenticationflow

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var ErrInvalidOTP = errors.New("invalid OTP")

func init() {
	RegisterIntent(&intentAuthenticate{})
	RegisterIntent(&intentLogin{})
	RegisterIntent(&intentSignup{})
	RegisterIntent(&intentAddLoginID{})
	RegisterIntent(&intentCreatePassword{})

	RegisterNode(&nodeCreatePassword{})
	RegisterNode(&nodeVerifyLoginID{})
	RegisterNode(&nodeLoginIDVerified{})
}

type intentAuthenticate struct {
	PretendLoginIDExists bool
}

var _ PublicFlow = &intentAuthenticate{}

func (*intentAuthenticate) Kind() string {
	return "intentAuthenticate"
}

func (*intentAuthenticate) FlowType() FlowType {
	return ""
}

func (*intentAuthenticate) FlowInit(r FlowReference, startFrom jsonpointer.T) {}

func (*intentAuthenticate) FlowFlowReference() FlowReference {
	return FlowReference{}
}

func (*intentAuthenticate) FlowRootObject(deps *Dependencies) (config.AuthenticationFlowObject, error) {
	return nil, nil
}

func (i *intentAuthenticate) CanReactTo(ctx context.Context, deps *Dependencies, flows Flows) (InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 {
		return &InputIntentAuthenticate{}, nil
	}

	return nil, ErrEOF
}

func (i *intentAuthenticate) ReactTo(ctx context.Context, deps *Dependencies, flows Flows, input Input) (*Node, error) {
	var inputLoginID InputLoginID

	switch {
	case AsInput(input, &inputLoginID):
		if i.PretendLoginIDExists {
			return NewSubFlow(&intentLogin{
				LoginID: inputLoginID.GetLoginID(),
			}), nil
		}

		return NewSubFlow(&intentSignup{
			LoginID: inputLoginID.GetLoginID(),
		}), nil
	default:
		return nil, ErrIncompatibleInput
	}
}

type intentLogin struct {
	LoginID string
}

var _ Intent = &intentLogin{}

func (*intentLogin) Kind() string {
	return "intentLogin"
}

func (*intentLogin) CanReactTo(ctx context.Context, deps *Dependencies, flows Flows) (InputSchema, error) {
	return nil, ErrEOF
}

func (*intentLogin) ReactTo(ctx context.Context, deps *Dependencies, flows Flows, input Input) (*Node, error) {
	return nil, ErrIncompatibleInput
}

type intentSignup struct {
	LoginID string
}

var _ Intent = &intentSignup{}

func (*intentSignup) Kind() string {
	return "intentSignup"
}

func (i *intentSignup) CanReactTo(ctx context.Context, deps *Dependencies, flows Flows) (InputSchema, error) {
	if len(flows.Nearest.Nodes) >= 2 {
		return nil, ErrEOF
	}

	return &InputIntentSignup{}, nil
}

func (i *intentSignup) ReactTo(ctx context.Context, deps *Dependencies, flows Flows, input Input) (*Node, error) {
	var inputLoginID InputLoginID
	var passwordInput InputCreatePasswordFlow

	switch {
	case AsInput(input, &inputLoginID) && inputLoginID.GetLoginID() != "":
		return NewSubFlow(&intentAddLoginID{
			LoginID: i.LoginID,
		}), nil
	case AsInput(input, &passwordInput) && passwordInput.IsCreatePassword():
		// In actual case, we check if the new password is valid against the password policy.
		return NewSubFlow(&intentCreatePassword{}), nil
	default:
		return nil, ErrIncompatibleInput
	}
}

type intentAddLoginID struct {
	LoginID string
}

var _ Intent = &intentAddLoginID{}

func (*intentAddLoginID) Kind() string {
	return "intentAddLoginID"
}

func (i *intentAddLoginID) CanReactTo(ctx context.Context, deps *Dependencies, flows Flows) (InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 {
		return &InputIntentAddLoginID{}, nil
	}

	return nil, ErrEOF
}

func (i *intentAddLoginID) ReactTo(ctx context.Context, deps *Dependencies, flows Flows, input Input) (*Node, error) {
	var inputLoginID InputLoginID

	switch {
	case AsInput(input, &inputLoginID):
		return NewNodeSimple(&nodeVerifyLoginID{
			LoginID: i.LoginID,
			OTP:     "123456",
		}), nil
	default:
		return nil, ErrIncompatibleInput
	}
}

type nodeVerifyLoginID struct {
	LoginID string
	OTP     string
}

var _ NodeSimple = &nodeVerifyLoginID{}

func (*nodeVerifyLoginID) Kind() string {
	return "nodeVerifyLoginID"
}

func (n *nodeVerifyLoginID) CanReactTo(ctx context.Context, deps *Dependencies, flows Flows) (InputSchema, error) {
	return &InputNodeVerifyLoginID{}, nil
}

func (n *nodeVerifyLoginID) ReactTo(ctx context.Context, deps *Dependencies, flows Flows, input Input) (*Node, error) {
	var otpInput InputOTP
	var resendInput InputResendOTP

	switch {
	case AsInput(input, &otpInput) && otpInput.GetOTP() != "":
		if n.OTP != otpInput.GetOTP() {
			return nil, ErrInvalidOTP
		}

		return NewNodeSimple(&nodeLoginIDVerified{
			LoginID: n.LoginID,
		}), nil
	case AsInput(input, &resendInput) && resendInput.IsResend():
		return NewNodeSimple(&nodeVerifyLoginID{
			LoginID: n.LoginID,
			OTP:     "654321",
		}), ErrUpdateNode
	default:
		return nil, ErrIncompatibleInput
	}
}

type nodeLoginIDVerified struct {
	LoginID string
}

var _ NodeSimple = &nodeLoginIDVerified{}

func (*nodeLoginIDVerified) Kind() string {
	return "nodeLoginIDVerified"
}

type intentCreatePassword struct{}

var _ Intent = &intentCreatePassword{}

func (*intentCreatePassword) Kind() string {
	return "intentCreatePassword"
}

func (*intentCreatePassword) CanReactTo(ctx context.Context, deps *Dependencies, flows Flows) (InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 {
		return &InputIntentCreatePassword{}, nil
	}

	return nil, ErrEOF
}

func (*intentCreatePassword) ReactTo(ctx context.Context, deps *Dependencies, flows Flows, input Input) (*Node, error) {
	var inputNewPassword InputNewPassword
	switch {
	case AsInput(input, &inputNewPassword):
		// Assume the new password fulfil the policy.
		return NewNodeSimple(&nodeCreatePassword{
			HashedNewPassword: inputNewPassword.GetNewPassword(),
		}), nil
	default:
		return nil, ErrIncompatibleInput
	}
}

type nodeCreatePassword struct {
	HashedNewPassword string
}

var _ NodeSimple = &nodeCreatePassword{}

func (*nodeCreatePassword) Kind() string {
	return "nodeCreatePassword"
}

type InputLoginID interface {
	GetLoginID() string
}

type InputOTP interface {
	GetOTP() string
}

type InputResendOTP interface {
	IsResend() bool
}

type InputCreatePasswordFlow interface {
	IsCreatePassword() bool
}

type InputNewPassword interface {
	GetNewPassword() string
}

type InputIntentAuthenticate struct {
	LoginID string `json:"login_id"`
}

var _ InputSchema = &InputIntentAuthenticate{}
var _ Input = &InputIntentAuthenticate{}
var _ InputLoginID = &InputIntentAuthenticate{}

func (*InputIntentAuthenticate) GetJSONPointer() jsonpointer.T {
	return nil
}

func (i *InputIntentAuthenticate) GetFlowRootObject() config.AuthenticationFlowObject {
	return nil
}

func (*InputIntentAuthenticate) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("login_id")
	b.Properties().Property("login_id", validation.SchemaBuilder{}.Type(validation.TypeString))
	return b
}

func (i *InputIntentAuthenticate) MakeInput(rawMessage json.RawMessage) (Input, error) {
	var input InputIntentAuthenticate
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func (*InputIntentAuthenticate) Input() {}

func (i *InputIntentAuthenticate) GetLoginID() string {
	return i.LoginID
}

type InputIntentSignup struct {
	LoginID        string `json:"login_id,omitempty"`
	CreatePassword bool   `json:"create_password,omitempty"`
}

var _ InputSchema = &InputIntentSignup{}
var _ Input = &InputIntentSignup{}
var _ InputLoginID = &InputIntentSignup{}
var _ InputCreatePasswordFlow = &InputIntentSignup{}

func (*InputIntentSignup) GetJSONPointer() jsonpointer.T {
	return nil
}

func (i *InputIntentSignup) GetFlowRootObject() config.AuthenticationFlowObject {
	return nil
}

func (*InputIntentSignup) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject)

	loginID := validation.SchemaBuilder{}.
		Required("login_id")
	loginID.Properties().Property("login_id", validation.SchemaBuilder{}.Type(validation.TypeString))

	createPassword := validation.SchemaBuilder{}.
		Required("create_password")
	createPassword.Properties().Property("create_password", validation.SchemaBuilder{}.Type(validation.TypeBoolean))

	b.OneOf(loginID, createPassword)

	return b
}

func (i *InputIntentSignup) MakeInput(rawMessage json.RawMessage) (Input, error) {
	var input InputIntentSignup
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func (*InputIntentSignup) Input() {}

func (i *InputIntentSignup) GetLoginID() string {
	return i.LoginID
}

func (i *InputIntentSignup) IsCreatePassword() bool {
	return i.CreatePassword
}

type InputIntentAddLoginID struct {
	LoginID string `json:"login_id,omitempty"`
}

var _ InputSchema = &InputIntentAddLoginID{}
var _ Input = &InputIntentAddLoginID{}
var _ InputLoginID = &InputIntentAddLoginID{}

func (*InputIntentAddLoginID) GetJSONPointer() jsonpointer.T {
	return nil
}

func (i *InputIntentAddLoginID) GetFlowRootObject() config.AuthenticationFlowObject {
	return nil
}

func (*InputIntentAddLoginID) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("login_id")
	b.Properties().Property("login_id", validation.SchemaBuilder{}.Type(validation.TypeString))
	return b
}

func (i *InputIntentAddLoginID) MakeInput(rawMessage json.RawMessage) (Input, error) {
	var input InputIntentAddLoginID
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func (*InputIntentAddLoginID) Input() {}

func (i *InputIntentAddLoginID) GetLoginID() string {
	return i.LoginID
}

type InputNodeVerifyLoginID struct {
	OTP    string `json:"otp,omitempty"`
	Resend bool   `json:"resend,omitempty"`
}

var _ InputSchema = &InputNodeVerifyLoginID{}
var _ Input = &InputNodeVerifyLoginID{}
var _ InputOTP = &InputNodeVerifyLoginID{}
var _ InputResendOTP = &InputNodeVerifyLoginID{}

func (*InputNodeVerifyLoginID) GetJSONPointer() jsonpointer.T {
	return nil
}

func (i *InputNodeVerifyLoginID) GetFlowRootObject() config.AuthenticationFlowObject {
	return nil
}

func (*InputNodeVerifyLoginID) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject)

	otp := validation.SchemaBuilder{}.
		Required("otp")
	otp.Properties().Property("otp", validation.SchemaBuilder{}.Type(validation.TypeString))

	resend := validation.SchemaBuilder{}.
		Required("resend")
	resend.Properties().Property("resend", validation.SchemaBuilder{}.Type(validation.TypeBoolean))

	b.OneOf(otp, resend)

	return b
}

func (i *InputNodeVerifyLoginID) MakeInput(rawMessage json.RawMessage) (Input, error) {
	var input InputNodeVerifyLoginID
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func (*InputNodeVerifyLoginID) Input() {}

func (i *InputNodeVerifyLoginID) GetOTP() string {
	return i.OTP
}

func (i *InputNodeVerifyLoginID) IsResend() bool {
	return i.Resend
}

type InputIntentCreatePassword struct {
	NewPassword string `json:"new_password,omitempty"`
}

var _ InputSchema = &InputIntentCreatePassword{}
var _ Input = &InputIntentCreatePassword{}
var _ InputNewPassword = &InputIntentCreatePassword{}

func (*InputIntentCreatePassword) GetJSONPointer() jsonpointer.T {
	return nil
}

func (i *InputIntentCreatePassword) GetFlowRootObject() config.AuthenticationFlowObject {
	return nil
}

func (*InputIntentCreatePassword) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("new_password")
	b.Properties().Property("new_password", validation.SchemaBuilder{}.Type(validation.TypeString))
	return b
}

func (i *InputIntentCreatePassword) MakeInput(rawMessage json.RawMessage) (Input, error) {
	var input InputIntentCreatePassword
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func (*InputIntentCreatePassword) Input() {}

func (i *InputIntentCreatePassword) GetNewPassword() string {
	return i.NewPassword
}
