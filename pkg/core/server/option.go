package server

type Option struct {
	GearPathPrefix string
	IsAPIVersioned bool
}

func DefaultOption() Option {
	return Option{
		GearPathPrefix: "",
		IsAPIVersioned: false,
	}
}
