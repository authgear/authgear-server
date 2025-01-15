//go:build !authgeardev
// +build !authgeardev

package resource

import (
	"github.com/spf13/afero"
)

func (o *NewManagerWithDirOptions) MakeBuiltinFSByBuildTag() afero.Fs {
	return afero.NewBasePathFs(afero.FromIOFS{FS: o.BuiltinResourceFS}, o.BuiltinResourceFSRoot)
}
