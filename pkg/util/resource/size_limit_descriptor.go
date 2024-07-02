package resource

type SizeLimitDescriptor interface {
	Descriptor
	GetSizeLimit() int // Size limit in bytes
}
