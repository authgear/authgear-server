package cloudstorage

type AssetItem struct {
	AssetID string `json:"asset_id"`
	URL     string `json:"url"`
}

type SignRequest struct {
	Assets []AssetItem `json:"assets"`
}
