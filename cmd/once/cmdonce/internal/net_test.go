package internal

import (
	"fmt"
	"net"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCheckTCPPortIsListening(t *testing.T) {
	Convey("CheckTCPPortIsListening", t, func() {
		Convey("When port is available", func() {
			// Find a free port by opening and immediately closing a listener
			//nolint: gosec // G102
			listener, err := net.Listen("tcp", ":0")
			if err != nil {
				t.Fatalf("Cannot set up test: %v", err)
			}

			addr := listener.Addr().(*net.TCPAddr)
			port := addr.Port
			listener.Close() // Close immediately to free the port

			// Now check if this port is available
			err = CheckTCPPortIsListening(port)
			So(err, ShouldBeNil)
		})

		Convey("When port is already in use", func() {
			// Find a free port by letting the OS assign one
			//nolint: gosec // G102
			listener, err := net.Listen("tcp", ":0")
			if err != nil {
				t.Fatalf("Cannot set up test: %v", err)
			}

			addr := listener.Addr().(*net.TCPAddr)
			port := addr.Port
			defer listener.Close() // Ensure the listener is closed after the test

			// The port should not be available now
			err = CheckTCPPortIsListening(port)
			So(err, ShouldHaveSameTypeAs, &ErrTCPPortAlreadyListening{})
			tcpErr, ok := err.(*ErrTCPPortAlreadyListening)
			So(ok, ShouldBeTrue)
			So(tcpErr.Port, ShouldEqual, port)
			So(err.Error(), ShouldContainSubstring, fmt.Sprintf("%d", port))
		})

		Convey("When binding to specific network types", func() {
			// This test simulates the scenario mentioned in the function comments
			// where a port might be available on one protocol but not another

			// Let OS find an available port
			//nolint: gosec // G102
			listener, err := net.Listen("tcp4", ":0")
			if err != nil {
				t.Skipf("Cannot set up test with tcp4: %v", err)
			}

			addr := listener.Addr().(*net.TCPAddr)
			port := addr.Port
			defer listener.Close()

			// Since we've bound to IPv4 specifically, the function should detect this
			// and return an error
			err = CheckTCPPortIsListening(port)
			So(err, ShouldHaveSameTypeAs, &ErrTCPPortAlreadyListening{})
		})

		Convey("With invalid port number", func() {
			// Go will accept this port but it's actually invalid
			// Port numbers should be between 0-65535
			port := 999999
			err := CheckTCPPortIsListening(port)
			So(err, ShouldHaveSameTypeAs, &ErrTCPPortAlreadyListening{})
		})

		Convey("With port at boundary values", func() {
			// Test with port 0 (let system assign)
			port := 0
			err := CheckTCPPortIsListening(port)
			So(err, ShouldBeNil)

			// Test with highest valid port 65535
			// First check if it's already in use
			//nolint: gosec // G102
			listener, err := net.Listen("tcp", ":65535")
			if err != nil {
				// Port is in use or not available, skip this check
				t.Logf("Port 65535 is not available for testing: %v", err)
			} else {
				listener.Close()
				err = CheckTCPPortIsListening(65535)
				So(err, ShouldBeNil)
			}
		})

		Convey("With negative port number", func() {
			// Negative port numbers should be rejected
			port := -1
			err := CheckTCPPortIsListening(port)
			So(err, ShouldHaveSameTypeAs, &ErrTCPPortAlreadyListening{})
		})
	})
}
