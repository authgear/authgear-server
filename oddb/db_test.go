package oddb

import (
	"testing"
)

type fakeConn struct {
	Conn
	OptionString string
}

type fakeDriver struct {
	Driver
}

func (driver fakeDriver) Open(optionString string) (Conn, error) {
	return fakeConn{
		OptionString: optionString,
	}, nil
}

func TestOpen(t *testing.T) {
	defer unregisterAllDrivers()

	Register("fakeImpl", fakeDriver{})

	if driver, err := Open("fakeImpl", "fakeOption"); err != nil {
		t.Fatalf("got err: %v, want a driver", err.Error())
	} else {
		if driver, ok := driver.(fakeConn); !ok {
			t.Fatalf("got driver = %v, want a driver of type fakeDriver", driver)
		} else {
			if driver.OptionString != "fakeOption" {
				t.Fatalf("got driver.OptionString = %v, want \"fakeOption\"", driver.OptionString)
			}
		}
	}
}
