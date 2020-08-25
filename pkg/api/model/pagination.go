package model

type PageCursor string

type PageItem struct {
	Value  interface{}
	Cursor PageCursor
}
