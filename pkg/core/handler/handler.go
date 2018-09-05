package handler

type Handler interface {
	Handle(Context)
}

type HandlerFunc func(Context)

func (f HandlerFunc) Handle(ctx Context) {
	f(ctx)
}
