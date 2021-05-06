package elasticsearch

type SortDirection string

const (
	SortDirectionDefault SortDirection = ""
	SortDirectionAsc     SortDirection = "asc"
	SortDirectionDesc    SortDirection = "desc"
)

type Stats struct {
	TotalCount int
}
