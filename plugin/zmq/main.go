// +build !zmq

package zmq

func init() {
	panic("zmq transport is not included during compilation, please compile with `go build -tags zmq`.")
}
