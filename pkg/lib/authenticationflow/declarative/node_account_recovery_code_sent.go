package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeAccountRecoveryCodeSent{})
}

type NodeAccountRecoveryCodeSent struct {
	TargetLoginID string `json:"target_login_id,omitempty"`
}

var _ authflow.NodeSimple = &NodeAccountRecoveryCodeSent{}

func (*NodeAccountRecoveryCodeSent) Kind() string {
	return "NodeAccountRecoveryCodeSent"
}
