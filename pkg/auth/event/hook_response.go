package event

import "github.com/skygeario/skygear-server/pkg/core/skyerr"

type HookResponse struct {
	IsAllowed bool        `json:"is_allowed"`
	Reason    string      `json:"reason"`
	Data      interface{} `json:"data"`
	Mutations *Mutations  `json:"mutations"`
}

func (resp HookResponse) Validate() error {
	if resp.IsAllowed {
		if resp.Mutations != nil {
			return skyerr.NewInvalidArgument("mutations must not exist", []string{"mutations"})
		}
		if resp.Reason != "" {
			return skyerr.NewInvalidArgument("reason must not exist", []string{"reason"})
		}
	} else {
		if resp.Reason == "" {
			return skyerr.NewInvalidArgument("reason must be provided", []string{"mutations"})
		}
	}
	return nil
}
