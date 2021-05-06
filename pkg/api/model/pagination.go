package model

type PageCursor string

type PageItemRef struct {
	ID     string
	Cursor PageCursor
}
