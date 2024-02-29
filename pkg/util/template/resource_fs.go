package template

import "github.com/authgear/authgear-server/pkg/util/resource"

type FsFilter = func(fs resource.Fs) bool

var AnyFs = func(fs resource.Fs) bool {
	return true
}

var ExcludeAppFs = func(fs resource.Fs) bool {
	if fs.GetFsLevel() == resource.FsLevelApp {
		return false
	}
	return true
}
