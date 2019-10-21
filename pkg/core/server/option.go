package server

type Option struct {
	RecoverPanic   bool
	GearPathPrefix string
}

func DefaultOption() Option {
	return Option{
		RecoverPanic:   true,
		GearPathPrefix: "",
	}
}
