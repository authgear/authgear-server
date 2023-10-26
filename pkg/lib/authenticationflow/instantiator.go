package authenticationflow

import "context"

type Instantiator interface {
	Instantiate(ctx context.Context, deps *Dependencies, flows Flows) error
}
