package handler

type HandlerFactory interface {
	NewHandler() Handler
}
