package handler

type RequestPayload interface {
	Validate() error
}

type EmptyRequestPayload struct{}

func (p EmptyRequestPayload) Validate() error {
	return nil
}
