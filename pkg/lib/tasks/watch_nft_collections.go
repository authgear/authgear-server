package tasks

const WatchNFTCollections = "WatchNFTCollections"

type WatchNFTCollectionsParam struct {
	AppID string
}

func (p *WatchNFTCollectionsParam) TaskName() string {
	return WatchNFTCollections
}
