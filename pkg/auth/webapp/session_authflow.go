package webapp

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/util/base32"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
)

type Authflow struct {
	FlowID       string                    `json:"flow_id"`
	InitialState *AuthflowState            `json:"initial_state,omitempty"`
	AllStates    map[string]*AuthflowState `json:"all_states,omitempty"`
}

type AuthflowState struct {
	XStep      string `json:"x_step"`
	StateToken string `json:"state_token"`
}

func newXStep() string {
	const (
		idAlphabet string = base32.Alphabet
		idLength   int    = 32
	)
	return corerand.StringWithAlphabet(idLength, idAlphabet, corerand.SecureRand)
}

func NewAuthflow(flowResponse *authflow.FlowResponse) *Authflow {
	state := &AuthflowState{
		XStep:      newXStep(),
		StateToken: flowResponse.StateToken,
	}

	return &Authflow{
		FlowID:       flowResponse.ID,
		InitialState: state,
		AllStates: map[string]*AuthflowState{
			state.XStep: state,
		},
	}
}

func (f *Authflow) Add(flowResponse *authflow.FlowResponse) {
	state := &AuthflowState{
		XStep:      newXStep(),
		StateToken: flowResponse.StateToken,
	}
	f.AllStates[state.XStep] = state
}
