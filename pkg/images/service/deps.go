package service

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewImagesCloudStorageServiceHTTPClient,
	wire.Struct(new(ImagesCloudStorageService), "*"),
)
