package transport

import (
	"net/url"
	"strconv"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

func requireParam(q url.Values, name string) (string, error) {
	s := q.Get(name)
	if s == "" {
		return "", apierrors.NewBadRequest(name + " is required")
	}
	return s, nil
}

func getIntParam(q url.Values, name string) (int, error) {
	s, err := requireParam(q, name)
	if err != nil {
		return 0, err
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, apierrors.NewBadRequest("invalid " + name + ": must be an integer")
	}
	return v, nil
}

func getDateParam(q url.Values, name string) (string, error) {
	s, err := requireParam(q, name)
	if err != nil {
		return "", err
	}
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return "", apierrors.NewBadRequest("invalid " + name + ": must be YYYY-MM-DD")
	}
	return s, nil
}

func validateMonth(name string, v int) error {
	if v < 1 || v > 12 {
		return apierrors.NewBadRequest("invalid " + name + ": must be between 1 and 12")
	}
	return nil
}
