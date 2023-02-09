package workflow

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

var ErrInvalidOTP = errors.New("invalid OTP")

func init() {
	RegisterPrivateIntent(&intentAuthenticate{})
	RegisterPrivateIntent(&intentLogin{})
	RegisterPrivateIntent(&intentSignup{})
	RegisterPrivateIntent(&intentAddLoginID{})
	RegisterPrivateIntent(&intentCreatePassword{})
	RegisterPrivateIntent(&intentFinishSignup{})

	RegisterNode(&nodeCreatePassword{})
	RegisterNode(&nodeVerifyLoginID{})
	RegisterNode(&nodeLoginIDVerified{})
}

type intentAuthenticate struct {
	PretendLoginIDExists bool
}

func (*intentAuthenticate) Kind() string {
	return "intentAuthenticate"
}

func (i *intentAuthenticate) JSONSchema() *validation.SimpleSchema {
	return EmptyJSONSchema
}

func (*intentAuthenticate) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (i *intentAuthenticate) CanReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Input, error) {
	if len(workflow.Nodes) == 0 {
		return []Input{
			&inputLoginID{},
		}, nil
	}

	return nil, ErrEOF
}

func (i *intentAuthenticate) ReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
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

func (i *intentAuthenticate) OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error) {
	return nil, nil
}

type InputLoginID interface {
	GetLoginID() string
}

type inputLoginID struct {
	LoginID string
}

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

func (*intentLogin) Kind() string {
	return "intentLogin"
}

func (*intentLogin) JSONSchema() *validation.SimpleSchema {
	return EmptyJSONSchema
}

func (*intentLogin) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (*intentLogin) CanReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Input, error) {
	return nil, ErrEOF
}

func (*intentLogin) ReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
	return nil, ErrIncompatibleInput
}

func (i *intentLogin) OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error) {
	return nil, nil
}

type intentSignup struct {
	LoginID string
}

func (*intentSignup) Kind() string {
	return "intentSignup"
}

func (*intentSignup) JSONSchema() *validation.SimpleSchema {
	return EmptyJSONSchema
}

func (*intentSignup) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (i *intentSignup) CanReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Input, error) {
	if len(workflow.Nodes) > 0 {
		lastNode := workflow.Nodes[len(workflow.Nodes)-1]
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

func (i *intentSignup) ReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
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

func (i *intentSignup) OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error) {
	return nil, nil
}

type intentAddLoginID struct {
	LoginID string
}

func (*intentAddLoginID) Kind() string {
	return "intentAddLoginID"
}

func (*intentAddLoginID) JSONSchema() *validation.SimpleSchema {
	return EmptyJSONSchema
}

func (*intentAddLoginID) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (i *intentAddLoginID) CanReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Input, error) {
	if len(workflow.Nodes) == 0 {
		return []Input{
			&inputLoginID{},
		}, nil
	}

	return nil, ErrEOF
}

func (i *intentAddLoginID) ReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
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

func (*intentAddLoginID) OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error) {
	return nil, nil
}

type nodeVerifyLoginID struct {
	LoginID string
	OTP     string
}

func (*nodeVerifyLoginID) Kind() string {
	return "nodeVerifyLoginID"
}

func (*nodeVerifyLoginID) GetEffects(ctx context.Context, deps *Dependencies, w *Workflow) ([]Effect, error) {
	return nil, nil
}

func (n *nodeVerifyLoginID) CanReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Input, error) {
	return []Input{
		&inputOTP{},
		&inputResendOTP{},
	}, nil
}

func (n *nodeVerifyLoginID) ReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
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

func (*nodeVerifyLoginID) OutputData(ctx context.Context, deps *Dependencies, w *Workflow) (interface{}, error) {
	return nil, nil
}

type InputOTP interface {
	GetOTP() string
}

type inputOTP struct {
	OTP string
}

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

func (*nodeLoginIDVerified) Kind() string {
	return "nodeLoginIDVerified"
}

func (*nodeLoginIDVerified) GetEffects(ctx context.Context, deps *Dependencies, w *Workflow) ([]Effect, error) {
	// In actual case, we create the identity here.
	return nil, nil
}

func (*nodeLoginIDVerified) CanReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Input, error) {
	// This workflow ends here.
	return nil, ErrEOF
}

func (*nodeLoginIDVerified) ReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
	return nil, ErrIncompatibleInput
}

func (*nodeLoginIDVerified) OutputData(ctx context.Context, deps *Dependencies, w *Workflow) (interface{}, error) {
	return nil, nil
}

type InputResendOTP interface {
	ResendOTP()
}

type inputResendOTP struct{}

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

func (*intentCreatePassword) Kind() string {
	return "intentCreatePassword"
}

func (*intentCreatePassword) JSONSchema() *validation.SimpleSchema {
	return EmptyJSONSchema
}

func (*intentCreatePassword) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (*intentCreatePassword) CanReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Input, error) {
	if len(workflow.Nodes) == 0 {
		return []Input{
			&inputNewPassword{},
		}, nil
	}

	return nil, ErrEOF
}

func (*intentCreatePassword) ReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
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

func (*intentCreatePassword) OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error) {
	return nil, nil
}

type nodeCreatePassword struct {
	HashedNewPassword string
}

func (*nodeCreatePassword) Kind() string {
	return "nodeCreatePassword"
}

func (*nodeCreatePassword) GetEffects(ctx context.Context, deps *Dependencies, w *Workflow) ([]Effect, error) {
	// In actual case, we create the password authenticator here.
	return nil, nil
}

func (*nodeCreatePassword) CanReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Input, error) {
	return nil, ErrEOF
}

func (*nodeCreatePassword) ReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
	return nil, ErrIncompatibleInput
}

func (*nodeCreatePassword) OutputData(ctx context.Context, deps *Dependencies, w *Workflow) (interface{}, error) {
	return nil, nil
}

type InputFinishSignup interface {
	FinishSignup()
}

type inputFinishSignup struct{}

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

func (*intentFinishSignup) Kind() string {
	return "intentFinishSignup"
}

func (*intentFinishSignup) JSONSchema() *validation.SimpleSchema {
	return EmptyJSONSchema
}

func (*intentFinishSignup) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (*intentFinishSignup) CanReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Input, error) {
	// In actual case, we have a lot to do in this workflow.
	// We have to check if the user has required identity, authenticator, 2FA set up.
	// And create session.
	return nil, ErrEOF
}

func (*intentFinishSignup) ReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
	return nil, ErrIncompatibleInput
}

func (*intentFinishSignup) OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error) {
	return nil, nil
}
