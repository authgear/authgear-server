package workflow2

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

// EmptyJSONSchema always validate successfully.
var EmptyJSONSchema = validation.NewSimpleSchema(`{}`)

var ErrInvalidOTP = errors.New("invalid OTP")

func init() {
	RegisterIntent(&intentAuthenticate{})
	RegisterIntent(&intentLogin{})
	RegisterIntent(&intentSignup{})
	RegisterIntent(&intentAddLoginID{})
	RegisterIntent(&intentCreatePassword{})
	RegisterIntent(&intentFinishSignup{})

	RegisterNode(&nodeCreatePassword{})
	RegisterNode(&nodeVerifyLoginID{})
	RegisterNode(&nodeLoginIDVerified{})
}

type intentAuthenticate struct {
	PretendLoginIDExists bool
}

var _ Intent = &intentAuthenticate{}

func (*intentAuthenticate) Kind() string {
	return "intentAuthenticate"
}

func (i *intentAuthenticate) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) ([]Input, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return []Input{
			&inputLoginID{},
		}, nil
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

type InputLoginID interface {
	GetLoginID() string
}

type inputLoginID struct {
	LoginID string
}

var _ Input = &inputLoginID{}

func (*inputLoginID) Kind() string {
	return "inputLoginID"
}

func (*inputLoginID) JSONSchema() *validation.SimpleSchema {
	return EmptyJSONSchema
}

func (i *inputLoginID) GetLoginID() string {
	return i.LoginID
}

type intentLogin struct {
	LoginID string
}

var _ Intent = &intentLogin{}

func (*intentLogin) Kind() string {
	return "intentLogin"
}

func (*intentLogin) JSONSchema() *validation.SimpleSchema {
	return EmptyJSONSchema
}

func (*intentLogin) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) ([]Input, error) {
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

func (*intentSignup) JSONSchema() *validation.SimpleSchema {
	return EmptyJSONSchema
}

func (i *intentSignup) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) ([]Input, error) {
	if len(workflows.Nearest.Nodes) > 0 {
		lastNode := workflows.Nearest.Nodes[len(workflows.Nearest.Nodes)-1]
		if lastNode.Type == NodeTypeSubWorkflow {
			intent := lastNode.SubWorkflow.Intent
			_, ok := intent.(*intentFinishSignup)
			if ok {
				return nil, ErrEOF
			}
		}
	}

	return []Input{
		&inputLoginID{},
		&inputCreatePasswordFlow{},
		&inputFinishSignup{},
	}, nil
}

func (i *intentSignup) ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error) {
	var inputLoginID InputLoginID
	var passwordInput InputCreatePasswordFlow
	var inputFinishSignup InputFinishSignup

	switch {
	case AsInput(input, &inputLoginID):
		return NewSubWorkflow(&intentAddLoginID{
			LoginID: i.LoginID,
		}), nil
	case AsInput(input, &passwordInput):
		// In actual case, we check if the new password is valid against the password policy.
		return NewSubWorkflow(&intentCreatePassword{}), nil
	case AsInput(input, &inputFinishSignup):
		return NewSubWorkflow(&intentFinishSignup{}), nil
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

func (*intentAddLoginID) JSONSchema() *validation.SimpleSchema {
	return EmptyJSONSchema
}

func (i *intentAddLoginID) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) ([]Input, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return []Input{
			&inputLoginID{},
		}, nil
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

func (n *nodeVerifyLoginID) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) ([]Input, error) {
	return []Input{
		&inputOTP{},
		&inputResendOTP{},
	}, nil
}

func (n *nodeVerifyLoginID) ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error) {
	var otpInput InputOTP
	var resendInput InputResendOTP

	switch {
	case AsInput(input, &otpInput):
		if n.OTP != otpInput.GetOTP() {
			return nil, ErrInvalidOTP
		}

		return NewNodeSimple(&nodeLoginIDVerified{
			LoginID: n.LoginID,
		}), nil
	case AsInput(input, &resendInput):
		return NewNodeSimple(&nodeVerifyLoginID{
			LoginID: n.LoginID,
			OTP:     "654321",
		}), ErrUpdateNode
	default:
		return nil, ErrIncompatibleInput
	}
}

type InputOTP interface {
	GetOTP() string
}

type inputOTP struct {
	OTP string
}

var _ Input = &inputOTP{}

func (*inputOTP) Kind() string {
	return "inputOTP"
}

func (i *inputOTP) JSONSchema() *validation.SimpleSchema {
	return EmptyJSONSchema
}

func (i *inputOTP) GetOTP() string {
	return i.OTP
}

type nodeLoginIDVerified struct {
	LoginID string
}

var _ NodeSimple = &nodeLoginIDVerified{}

func (*nodeLoginIDVerified) Kind() string {
	return "nodeLoginIDVerified"
}

type InputResendOTP interface {
	ResendOTP()
}

type inputResendOTP struct{}

var _ Input = &inputResendOTP{}

func (*inputResendOTP) Kind() string {
	return "inputResendOTP"
}

func (i *inputResendOTP) JSONSchema() *validation.SimpleSchema {
	return EmptyJSONSchema
}

func (*inputResendOTP) ResendOTP() {}

type InputCreatePasswordFlow interface {
	CreatePassword()
}

type inputCreatePasswordFlow struct{}

var _ Input = &inputCreatePasswordFlow{}

func (*inputCreatePasswordFlow) Kind() string {
	return "inputCreatePasswordFlow"
}

func (i *inputCreatePasswordFlow) JSONSchema() *validation.SimpleSchema {
	return EmptyJSONSchema
}

func (i *inputCreatePasswordFlow) CreatePassword() {}

type InputNewPassword interface {
	GetNewPassword() string
}

type inputNewPassword struct {
	NewPassword string
}

var _ Input = &inputNewPassword{}

func (*inputNewPassword) Kind() string {
	return "inputNewPassword"
}

func (*inputNewPassword) JSONSchema() *validation.SimpleSchema {
	return EmptyJSONSchema
}

func (i *inputNewPassword) GetNewPassword() string {
	return i.NewPassword
}

type intentCreatePassword struct{}

var _ Intent = &intentCreatePassword{}

func (*intentCreatePassword) Kind() string {
	return "intentCreatePassword"
}

func (*intentCreatePassword) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) ([]Input, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return []Input{
			&inputNewPassword{},
		}, nil
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

type InputFinishSignup interface {
	FinishSignup()
}

type inputFinishSignup struct{}

var _ Input = &inputFinishSignup{}

func (*inputFinishSignup) Kind() string {
	return "inputFinishSignup"
}

func (i *inputFinishSignup) JSONSchema() *validation.SimpleSchema {
	return EmptyJSONSchema
}

func (i *inputFinishSignup) Instantiate(data json.RawMessage) error {
	return json.Unmarshal(data, i)
}

func (*inputFinishSignup) FinishSignup() {}

type intentFinishSignup struct{}

var _ Intent = &intentFinishSignup{}

func (*intentFinishSignup) Kind() string {
	return "intentFinishSignup"
}

func (*intentFinishSignup) CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) ([]Input, error) {
	// In actual case, we have a lot to do in this workflow.
	// We have to check if the user has required identity, authenticator, 2FA set up.
	// And create session.
	return nil, ErrEOF
}

func (*intentFinishSignup) ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error) {
	return nil, ErrIncompatibleInput
}
