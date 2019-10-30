package cloudstorage

type ListObjectsRequest struct {
	Prefix          string
	PageSize        int
	PaginationToken string
}
