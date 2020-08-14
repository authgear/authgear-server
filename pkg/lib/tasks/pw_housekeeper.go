package tasks

import (
	"errors"
)

const PwHousekeeper = "PwHousekeeper"

type PwHousekeeperParam struct {
	UserID string
}

func (p *PwHousekeeperParam) Validate() error {
	if p.UserID == "" {
		return errors.New("missing user ID")
	}

	return nil
}

func (p *PwHousekeeperParam) TaskName() string {
	return PwHousekeeper
}
