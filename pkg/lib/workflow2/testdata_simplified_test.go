package workflow2

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

var ErrInvalidOTP = errors.New("invalid OTP")

func init() {
	RegisterIntent(&intentAuthenticate{})
	RegisterIntent(&intentLogin{})
	RegisterIntent(&intentSignup{})
	RegisterIntent(&intentAddLoginID{})
	RegisterIntent(&intentCreatePassword{})

	RegisterIntent(&intentTestBoundarySteps{})
	RegisterIntent(&intentTestBoundaryStep{})

	RegisterNode(&nodeCreatePassword{})
	RegisterNode(&nodeVerifyLoginID{})
	RegisterNode(&nodeLoginIDVerified{})

	RegisterNode(&nodeTestBoundary{})
}

type intentAuthenticate struct {
	PretendLoginIDExists bool
}

var _ Intent = &intentAuthenticate{}

func (*intentAuthenticate) Kind() string {
	return "intentAuthenticate"
}

func (i *intentAuthenticate) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) (InputSchema, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return &InputIntentAuthenticate{}, nil
	}

	return nil, ErrEOF
}

func (i *intentAuthenticate) ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error) {
	var inputLoginID InputLoginID

	switch {
	case AsInput(input, &inputLoginID):
		if i.PretendLoginIDExists {
			return NewSubWorkflow(&intentLogin{
				LoginID: inputLoginID.GetLoginID(),
			}), nil
		}

		return NewSubWorkflow(&intentSignup{
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

func (*intentLogin) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) (InputSchema, error) {
	return nil, ErrEOF
}

func (*intentLogin) ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error) {
	return nil, ErrIncompatibleInput
}

type intentSignup struct {
	LoginID string
}

var _ Intent = &intentSignup{}

func (*intentSignup) Kind() string {
	return "intentSignup"
}

func (i *intentSignup) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) (InputSchema, error) {
	if len(workflows.Nearest.Nodes) >= 2 {
		return nil, ErrEOF
	}

	return &InputIntentSignup{}, nil
}

func (i *intentSignup) ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error) {
	var inputLoginID InputLoginID
	var passwordInput InputCreatePasswordFlow

	switch {
	case AsInput(input, &inputLoginID) && inputLoginID.GetLoginID() != "":
		return NewSubWorkflow(&intentAddLoginID{
			LoginID: i.LoginID,
		}), nil
	case AsInput(input, &passwordInput) && passwordInput.IsCreatePassword():
		// In actual case, we check if the new password is valid against the password policy.
		return NewSubWorkflow(&intentCreatePassword{}), nil
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

func (i *intentAddLoginID) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) (InputSchema, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return &InputIntentAddLoginID{}, nil
	}

	return nil, ErrEOF
}

func (i *intentAddLoginID) ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error) {
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

func (n *nodeVerifyLoginID) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) (InputSchema, error) {
	return &InputNodeVerifyLoginID{}, nil
}

func (n *nodeVerifyLoginID) ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error) {
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

func (*intentCreatePassword) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) (InputSchema, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return &InputIntentCreatePassword{}, nil
	}

	return nil, ErrEOF
}

func (*intentCreatePassword) ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error) {
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

func (*InputIntentAuthenticate) Input() {}

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

func (*InputIntentSignup) Input() {}

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

func (*InputIntentAddLoginID) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("login_id")
	b.Properties().Property("login_id", validation.SchemaBuilder{}.Type(validation.TypeString))
	return b
}

func (*InputIntentAddLoginID) Input() {}

func (i *InputIntentAddLoginID) MakeInput(rawMessage json.RawMessage) (Input, error) {
	var input InputIntentAddLoginID
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

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

type intentTestBoundarySteps struct{}

var _ Intent = &intentTestBoundarySteps{}

func (*intentTestBoundarySteps) Kind() string {
	return "intentTestBoundarySteps"
}

func (*intentTestBoundarySteps) CanReactTo(ctx context.Context, deps *Dependencies, workflow Workflows) (InputSchema, error) {
	return nil, nil
}

func (*intentTestBoundarySteps) ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error) {
	name := strconv.Itoa(len(workflows.Nearest.Nodes))
	return NewSubWorkflow(&intentTestBoundaryStep{
		Name: name,
	}), nil
}

type intentTestBoundaryStep struct {
	Name string
}

var _ Intent = &intentTestBoundaryStep{}
var _ Boundary = &intentTestBoundaryStep{}

func (*intentTestBoundaryStep) Kind() string {
	return "intentTestBoundaryStep"
}

func (i *intentTestBoundaryStep) Boundary() string {
	return i.Name
}

func (i *intentTestBoundaryStep) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) (InputSchema, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		return &inputTestBoundary{}, nil
	default:
		return nil, ErrEOF
	}
}

func (i *intentTestBoundaryStep) ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		return NewNodeSimple(&nodeTestBoundary{}), nil
	default:
		return nil, ErrIncompatibleInput
	}
}

type nodeTestBoundary struct{}

var _ NodeSimple = &nodeTestBoundary{}

func (*nodeTestBoundary) Kind() string {
	return "nodeTestBoundary"
}

type InputTestBoundary interface {
	InputTestBoundary()
}

type inputTestBoundary struct{}

var _ Input = &inputTestBoundary{}
var _ InputSchema = &inputTestBoundary{}
var _ InputTestBoundary = &inputTestBoundary{}

func (*inputTestBoundary) Input() {}
func (*inputTestBoundary) SchemaBuilder() validation.SchemaBuilder {
	return validation.SchemaBuilder{}
}
func (i *inputTestBoundary) MakeInput(rawMessage json.RawMessage) (Input, error) {
	var input inputTestBoundary
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}
func (i *inputTestBoundary) InputTestBoundary() {}
