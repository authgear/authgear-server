package cloudstorage

type ListObjectsResponse struct {
	// PaginationToken is the token to retrieve next page.
	// If it is absent, there is no next page.
	PaginationToken string      `json:"pagination_token,omitempty"`
	Assets          []AssetItem `json:"assets"`
}
