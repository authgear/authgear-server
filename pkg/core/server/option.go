package server

type Option struct {
	RecoverPanic bool
}

func DefaultOption() Option {
	return Option{
		RecoverPanic: true,
	}
}
