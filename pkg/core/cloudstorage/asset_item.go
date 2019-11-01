package cloudstorage

type ListAssetItem struct {
	AssetName string `json:"asset_name"`
	Size      int64  `json:"size,omitempty"`
}

type SignedAssetItem struct {
	AssetName string `json:"asset_name"`
	URL       string `json:"url,omitempty"`
}

type SignAssetItem struct {
	AssetName string `json:"asset_name"`
	Expire    int    `json:"expire,omitempty"`
}
