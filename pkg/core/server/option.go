package server

type Option struct {
	RecoverPanic   bool
	GearPathPrefix string
	IsAPIVersioned bool
}

func DefaultOption() Option {
	return Option{
		RecoverPanic:   true,
		GearPathPrefix: "",
		IsAPIVersioned: false,
	}
}
