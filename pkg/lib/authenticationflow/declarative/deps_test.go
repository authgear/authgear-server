package declarative_test

import "github.com/authgear/authgear-server/pkg/lib/authenticationflow"

//go:generate go tool mockgen -source=deps_test.go -destination=deps_mock_test.go -package declarative_test

type UserService interface {
	authenticationflow.UserService
}
