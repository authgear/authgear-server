package errors_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/errors"
)

func detailsA() error {
	return errors.New("error 1")
}

func detailsB() error {
	return errors.WithDetails(detailsA(), errors.Details{"data": 123})
}

func detailsC() error {
	return errors.HandledWithMessage(detailsB(), "error 2")
}

func detailsD() error {
	return errors.WithDetails(detailsC(), errors.Details{"data": 456, "value": "test"})
}

func TestDetails(t *testing.T) {
	Convey("WithDetails/CollectDetails", t, func() {
		err := detailsD()
		So(err, ShouldBeError, "error 2")
		details := errors.CollectDetails(err, nil)
		So(details, ShouldResemble, errors.Details{
			"data":  456,
			"value": "test",
		})
	})
}
