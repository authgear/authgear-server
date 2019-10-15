package cloudstorage

type AssetItem struct {
	AssetName string `json:"asset_name"`
	Size      int64  `json:"size,omitempty"`
	URL       string `json:"url,omitempty"`
}
