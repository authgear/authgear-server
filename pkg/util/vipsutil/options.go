package vipsutil

type Options struct {
	Width            int
	Height           int
	ResizingModeType ResizingModeType
}

func (o Options) ShouldResize() bool {
	return o.ResizingModeType != "" && o.Width > 0 && o.Height > 0
}
