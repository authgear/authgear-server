package workflow

import (
	"errors"
)

var ErrInvalidOTP = errors.New("invalid OTP")

type intentAuthenticate struct {
	PretendLoginIDExists bool
}

func (*intentAuthenticate) GetEffects(ctx *Context, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (i *intentAuthenticate) DeriveEdges(ctx *Context, workflow *Workflow) ([]Edge, error) {
	if len(workflow.Nodes) == 0 {
		return []Edge{
			&edgeTakeLoginID{
				PretendLoginIDExists: i.PretendLoginIDExists,
			},
		}, nil
	}

	return nil, ErrEOF
}

func (i *intentAuthenticate) OutputData(ctx *Context, workflow *Workflow) (interface{}, error) {
	return nil, nil
}

type edgeTakeLoginID struct {
	PretendLoginIDExists bool
}

func (e *edgeTakeLoginID) Instantiate(ctx *Context, workflow *Workflow, input interface{}) (*Node, error) {
	var inputLoginID InputLoginID
	ok := Input(input, &inputLoginID)
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

func (i *inputLoginID) GetLoginID() string {
	return i.LoginID
}

func (i *inputLoginID) AddLoginID() {}

type intentLogin struct {
	LoginID string
}

func (*intentLogin) GetEffects(ctx *Context, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (i *intentLogin) DeriveEdges(ctx *Context, workflow *Workflow) ([]Edge, error) {
	return nil, ErrEOF
}

func (i *intentLogin) OutputData(ctx *Context, workflow *Workflow) (interface{}, error) {
	return nil, nil
}

type intentSignup struct {
	LoginID string
}

func (*intentSignup) GetEffects(ctx *Context, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (i *intentSignup) DeriveEdges(ctx *Context, workflow *Workflow) ([]Edge, error) {
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

func (i *intentSignup) OutputData(ctx *Context, workflow *Workflow) (interface{}, error) {
	return nil, nil
}

type edgeAddLoginIDFlow struct {
	LoginID string
}

func (e *edgeAddLoginIDFlow) Instantiate(ctx *Context, workflow *Workflow, input interface{}) (*Node, error) {
	var inputAddLoginIDFlow InputAddLoginIDFlow
	ok := Input(input, &inputAddLoginIDFlow)
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

func (*intentAddLoginID) GetEffects(ctx *Context, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (i *intentAddLoginID) DeriveEdges(ctx *Context, workflow *Workflow) ([]Edge, error) {
	if len(workflow.Nodes) == 0 {
		return []Edge{
			&edgeVerifyLoginID{
				LoginID: i.LoginID,
			},
		}, nil
	}

	return nil, ErrEOF
}

func (*intentAddLoginID) OutputData(ctx *Context, workflow *Workflow) (interface{}, error) {
	return nil, nil
}

type edgeVerifyLoginID struct {
	LoginID string
}

func (e *edgeVerifyLoginID) Instantiate(ctx *Context, workflow *Workflow, input interface{}) (*Node, error) {
	return NewNodeSimple(&nodeVerifyLoginID{
		LoginID: e.LoginID,
		OTP:     "123456",
	}), nil
}

type nodeVerifyLoginID struct {
	LoginID string
	OTP     string
}

func (*nodeVerifyLoginID) GetEffects(ctx *Context) ([]Effect, error) {
	return nil, nil
}

func (n *nodeVerifyLoginID) DeriveEdges(ctx *Context) ([]Edge, error) {
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

func (*nodeVerifyLoginID) OutputData(ctx *Context) (interface{}, error) {
	return nil, nil
}

type edgeVerifyOTP struct {
	LoginID string
	OTP     string
}

func (e *edgeVerifyOTP) Instantiate(ctx *Context, workflow *Workflow, input interface{}) (*Node, error) {
	var otpInput InputOTP
	ok := Input(input, &otpInput)
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

func (i *inputOTP) GetOTP() string {
	return i.OTP
}

type nodeLoginIDVerified struct {
	LoginID string
}

func (*nodeLoginIDVerified) GetEffects(ctx *Context) ([]Effect, error) {
	// In actual case, we create the identity here.
	return nil, nil
}

func (*nodeLoginIDVerified) DeriveEdges(ctx *Context) ([]Edge, error) {
	// This workflow ends here.
	return nil, ErrEOF
}

func (*nodeLoginIDVerified) OutputData(ctx *Context) (interface{}, error) {
	return nil, nil
}

type edgeResendOTP struct {
	LoginID string
}

func (e *edgeResendOTP) Instantiate(ctx *Context, workflow *Workflow, input interface{}) (*Node, error) {
	var resendInput InputResendOTP
	ok := Input(input, &resendInput)
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

func (*inputResendOTP) ResendOTP() {}

type edgeCreatePasswordFlow struct{}

func (*edgeCreatePasswordFlow) Instantiate(ctx *Context, workflow *Workflow, input interface{}) (*Node, error) {
	var passwordInput InputCreatePasswordFlow
	ok := Input(input, &passwordInput)
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

func (i *inputCreatePasswordFlow) CreatePassword() {}

type InputNewPassword interface {
	GetNewPassword() string
}

type inputNewPassword struct {
	NewPassword string
}

func (i *inputNewPassword) GetNewPassword() string {
	return i.NewPassword
}

type intentCreatePassword struct{}

func (*intentCreatePassword) GetEffects(ctx *Context, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (*intentCreatePassword) DeriveEdges(ctx *Context, workflow *Workflow) ([]Edge, error) {
	if len(workflow.Nodes) == 0 {
		return []Edge{
			&edgeCheckPasswordAgainstPolicy{},
		}, nil
	}

	return nil, ErrEOF
}

func (*intentCreatePassword) OutputData(ctx *Context, workflow *Workflow) (interface{}, error) {
	return nil, nil
}

type edgeCheckPasswordAgainstPolicy struct{}

func (*edgeCheckPasswordAgainstPolicy) Instantiate(ctx *Context, workflow *Workflow, input interface{}) (*Node, error) {

	var inputNewPassword InputNewPassword
	ok := Input(input, &inputNewPassword)
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

func (*nodeCreatePassword) GetEffects(ctx *Context) ([]Effect, error) {
	// In actual case, we create the password authenticator here.
	return nil, nil
}

func (*nodeCreatePassword) DeriveEdges(ctx *Context) ([]Edge, error) {
	return nil, ErrEOF
}

func (*nodeCreatePassword) OutputData(ctx *Context) (interface{}, error) {
	return nil, nil
}

type edgeFinishSignup struct{}

func (*edgeFinishSignup) Instantiate(ctx *Context, workflow *Workflow, input interface{}) (*Node, error) {
	var inputFinishSignup InputFinishSignup
	ok := Input(input, &inputFinishSignup)
	if !ok {
		return nil, ErrIncompatibleInput
	}

	return NewSubWorkflow(&intentFinishSignup{}), nil
}

type InputFinishSignup interface {
	FinishSignup()
}

type inputFinishSignup struct{}

func (*inputFinishSignup) FinishSignup() {}

type intentFinishSignup struct{}

func (*intentFinishSignup) GetEffects(ctx *Context, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (*intentFinishSignup) DeriveEdges(ctx *Context, workflow *Workflow) ([]Edge, error) {
	// In actual case, we have a lot to do in this workflow.
	// We have to check if the user has required identity, authenticator, 2FA set up.
	// And create session.
	return nil, ErrEOF
}

func (*intentFinishSignup) OutputData(ctx *Context, workflow *Workflow) (interface{}, error) {
	return nil, nil
}
