package internal

import (
	"fmt"
	"net"
)

// CheckTCPPortIsListening checks if a TCP port can be listened to.
// Returns nil if the port is available (can be listened to),
// ErrTCPPortAlreadyListening if the port is already in use.
func CheckTCPPortIsListening(port int) error {
	address := fmt.Sprintf(":%v", port)

	// I discover that if the first argument is "tcp",
	// then no error is returned if it actually can not bind to IPv4 interfaces.
	//
	// This can be verified by running `nc -l 0.0.0.0 80`, which listens to IPv4 interfaces only.
	// net.Listen("tcp", ":80") WILL NOT return error.
	//
	// Therefore, in additional to "tcp", we also explicitly test for "tcp4" and "tcp6".
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return &ErrTCPPortAlreadyListening{
			Port: port,
		}
	}
	_ = listener.Close()

	listener, err = net.Listen("tcp4", address)
	if err != nil {
		return &ErrTCPPortAlreadyListening{
			Port: port,
		}
	}
	_ = listener.Close()

	listener, err = net.Listen("tcp6", address)
	if err != nil {
		return &ErrTCPPortAlreadyListening{
			Port: port,
		}
	}
	_ = listener.Close()

	return nil
}
