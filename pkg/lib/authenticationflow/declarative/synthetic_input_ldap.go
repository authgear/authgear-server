package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type SyntheticInputLDAP struct {
	ServerName string
	Username   string
	Password   string
}

var _ authflow.Input = &SyntheticInputLDAP{}
var _ inputTakeIdentificationMethod = &SyntheticInputOAuth{}
var _ inputTakeLDAP = &SyntheticInputLDAP{}

func (*SyntheticInputLDAP) Input() {}

func (i *SyntheticInputLDAP) GetIdentificationMethod() config.AuthenticationFlowIdentification {
	return config.AuthenticationFlowIdentificationLDAP
}

func (i *SyntheticInputLDAP) GetServerName() string {
	return i.ServerName
}

func (i *SyntheticInputLDAP) GetUsername() string {
	return i.Username
}

func (i *SyntheticInputLDAP) GetPassword() string {
	return i.Password
}
