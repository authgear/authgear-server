package workflow

import (
	"context"
	"errors"
)

var ErrInvalidOTP = errors.New("invalid OTP")

type intentAuthenticate struct {
	PretendLoginIDExists bool
}

func (*intentAuthenticate) Kind() string {
	return "intentAuthenticate"
}

func (*intentAuthenticate) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (i *intentAuthenticate) DeriveEdges(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Edge, error) {
	if len(workflow.Nodes) == 0 {
		return []Edge{
			&edgeTakeLoginID{
				PretendLoginIDExists: i.PretendLoginIDExists,
			},
		}, nil
	}

	return nil, ErrEOF
}

func (i *intentAuthenticate) OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error) {
	return nil, nil
}

type edgeTakeLoginID struct {
	PretendLoginIDExists bool
}

func (e *edgeTakeLoginID) Instantiate(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
	var inputLoginID InputLoginID
	ok := AsInput(input, &inputLoginID)
	if !ok {
		return nil, ErrIncompatibleInput
	}

	if e.PretendLoginIDExists {
		return NewSubWorkflow(&intentLogin{
			LoginID: inputLoginID.GetLoginID(),
		}), nil
	}

	return NewSubWorkflow(&intentSignup{
		LoginID: inputLoginID.GetLoginID(),
	}), nil
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

func (i *inputLoginID) GetLoginID() string {
	return i.LoginID
}

func (i *inputLoginID) AddLoginID() {}

type intentLogin struct {
	LoginID string
}

func (*intentLogin) Kind() string {
	return "intentLogin"
}

func (*intentLogin) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (i *intentLogin) DeriveEdges(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Edge, error) {
	return nil, ErrEOF
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

func (*intentSignup) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (i *intentSignup) DeriveEdges(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Edge, error) {
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

	return []Edge{
		&edgeAddLoginIDFlow{
			LoginID: i.LoginID,
		},
		&edgeCreatePasswordFlow{},
		&edgeFinishSignup{},
	}, nil
}

func (i *intentSignup) OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error) {
	return nil, nil
}

type edgeAddLoginIDFlow struct {
	LoginID string
}

func (e *edgeAddLoginIDFlow) Instantiate(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
	var inputAddLoginIDFlow InputAddLoginIDFlow
	ok := AsInput(input, &inputAddLoginIDFlow)
	if !ok {
		return nil, ErrIncompatibleInput
	}

	return NewSubWorkflow(&intentAddLoginID{
		LoginID: e.LoginID,
	}), nil
}

type InputAddLoginIDFlow interface {
	AddLoginID()
}

type intentAddLoginID struct {
	LoginID string
}

func (*intentAddLoginID) Kind() string {
	return "intentAddLoginID"
}

func (*intentAddLoginID) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (i *intentAddLoginID) DeriveEdges(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Edge, error) {
	if len(workflow.Nodes) == 0 {
		return []Edge{
			&edgeVerifyLoginID{
				LoginID: i.LoginID,
			},
		}, nil
	}

	return nil, ErrEOF
}

func (*intentAddLoginID) OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error) {
	return nil, nil
}

type edgeVerifyLoginID struct {
	LoginID string
}

func (e *edgeVerifyLoginID) Instantiate(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
	return NewNodeSimple(&nodeVerifyLoginID{
		LoginID: e.LoginID,
		OTP:     "123456",
	}), nil
}

type nodeVerifyLoginID struct {
	LoginID string
	OTP     string
}

func (*nodeVerifyLoginID) Kind() string {
	return "nodeVerifyLoginID"
}

func (*nodeVerifyLoginID) GetEffects(ctx context.Context, deps *Dependencies) ([]Effect, error) {
	return nil, nil
}

func (n *nodeVerifyLoginID) DeriveEdges(ctx context.Context, deps *Dependencies) ([]Edge, error) {
	return []Edge{
		&edgeVerifyOTP{
			LoginID: n.LoginID,
			OTP:     n.OTP,
		},
		&edgeResendOTP{
			LoginID: n.LoginID,
		},
	}, nil
}

func (*nodeVerifyLoginID) OutputData(ctx context.Context, deps *Dependencies) (interface{}, error) {
	return nil, nil
}

type edgeVerifyOTP struct {
	LoginID string
	OTP     string
}

func (e *edgeVerifyOTP) Instantiate(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
	var otpInput InputOTP
	ok := AsInput(input, &otpInput)
	if !ok {
		return nil, ErrIncompatibleInput
	}

	if e.OTP != otpInput.GetOTP() {
		return nil, ErrInvalidOTP
	}

	return NewNodeSimple(&nodeLoginIDVerified{
		LoginID: e.LoginID,
	}), nil
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

func (i *inputOTP) GetOTP() string {
	return i.OTP
}

type nodeLoginIDVerified struct {
	LoginID string
}

func (*nodeLoginIDVerified) Kind() string {
	return "nodeLoginIDVerified"
}

func (*nodeLoginIDVerified) GetEffects(ctx context.Context, deps *Dependencies) ([]Effect, error) {
	// In actual case, we create the identity here.
	return nil, nil
}

func (*nodeLoginIDVerified) DeriveEdges(ctx context.Context, deps *Dependencies) ([]Edge, error) {
	// This workflow ends here.
	return nil, ErrEOF
}

func (*nodeLoginIDVerified) OutputData(ctx context.Context, deps *Dependencies) (interface{}, error) {
	return nil, nil
}

type edgeResendOTP struct {
	LoginID string
}

func (e *edgeResendOTP) Instantiate(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
	var resendInput InputResendOTP
	ok := AsInput(input, &resendInput)
	if !ok {
		return nil, ErrIncompatibleInput
	}

	return NewNodeSimple(&nodeVerifyLoginID{
		LoginID: e.LoginID,
		OTP:     "654321",
	}), ErrUpdateNode
}

type InputResendOTP interface {
	ResendOTP()
}

type inputResendOTP struct{}

func (*inputResendOTP) Kind() string {
	return "inputResendOTP"
}

func (*inputResendOTP) ResendOTP() {}

type edgeCreatePasswordFlow struct{}

func (*edgeCreatePasswordFlow) Instantiate(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
	var passwordInput InputCreatePasswordFlow
	ok := AsInput(input, &passwordInput)
	if !ok {
		return nil, ErrIncompatibleInput
	}

	// In actual case, we check if the new password is valid against the password policy.
	return NewSubWorkflow(&intentCreatePassword{}), nil
}

type InputCreatePasswordFlow interface {
	CreatePassword()
}

type inputCreatePasswordFlow struct{}

func (*inputCreatePasswordFlow) Kind() string {
	return "inputCreatePasswordFlow"
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

func (i *inputNewPassword) GetNewPassword() string {
	return i.NewPassword
}

type intentCreatePassword struct{}

func (*intentCreatePassword) Kind() string {
	return "intentCreatePassword"
}

func (*intentCreatePassword) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (*intentCreatePassword) DeriveEdges(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Edge, error) {
	if len(workflow.Nodes) == 0 {
		return []Edge{
			&edgeCheckPasswordAgainstPolicy{},
		}, nil
	}

	return nil, ErrEOF
}

func (*intentCreatePassword) OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error) {
	return nil, nil
}

type edgeCheckPasswordAgainstPolicy struct{}

func (*edgeCheckPasswordAgainstPolicy) Instantiate(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {

	var inputNewPassword InputNewPassword
	ok := AsInput(input, &inputNewPassword)
	if !ok {
		return nil, ErrIncompatibleInput
	}

	// Assume the new password fulfil the policy.
	return NewNodeSimple(&nodeCreatePassword{
		HashedNewPassword: inputNewPassword.GetNewPassword(),
	}), nil
}

type nodeCreatePassword struct {
	HashedNewPassword string
}

func (*nodeCreatePassword) Kind() string {
	return "nodeCreatePassword"
}

func (*nodeCreatePassword) GetEffects(ctx context.Context, deps *Dependencies) ([]Effect, error) {
	// In actual case, we create the password authenticator here.
	return nil, nil
}

func (*nodeCreatePassword) DeriveEdges(ctx context.Context, deps *Dependencies) ([]Edge, error) {
	return nil, ErrEOF
}

func (*nodeCreatePassword) OutputData(ctx context.Context, deps *Dependencies) (interface{}, error) {
	return nil, nil
}

type edgeFinishSignup struct{}

func (*edgeFinishSignup) Instantiate(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
	var inputFinishSignup InputFinishSignup
	ok := AsInput(input, &inputFinishSignup)
	if !ok {
		return nil, ErrIncompatibleInput
	}

	return NewSubWorkflow(&intentFinishSignup{}), nil
}

type InputFinishSignup interface {
	FinishSignup()
}

type inputFinishSignup struct{}

func (*inputFinishSignup) Kind() string {
	return "inputFinishSignup"
}

func (*inputFinishSignup) FinishSignup() {}

type intentFinishSignup struct{}

func (*intentFinishSignup) Kind() string {
	return "intentFinishSignup"
}

func (*intentFinishSignup) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (*intentFinishSignup) DeriveEdges(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Edge, error) {
	// In actual case, we have a lot to do in this workflow.
	// We have to check if the user has required identity, authenticator, 2FA set up.
	// And create session.
	return nil, ErrEOF
}

func (*intentFinishSignup) OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error) {
	return nil, nil
}
