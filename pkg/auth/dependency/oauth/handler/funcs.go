package handler

type ScopesValidator func(scopes []string) error
type TokenGenerator func()string
