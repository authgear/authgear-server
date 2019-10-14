package cloudstorage

type AssetItem struct {
	AssetName string `json:"asset_name"`
	URL       string `json:"url"`
}

type SignRequest struct {
	Assets []AssetItem `json:"assets"`
}
