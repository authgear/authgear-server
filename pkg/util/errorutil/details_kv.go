package errorutil

import "fmt"

type DetailTag string

const (
	SafeDetail DetailTag = "safe"
)

func (t DetailTag) Value(value any) DetailTaggedValue {
	return DetailTaggedValue{t, value}
}

type DetailTaggedValue struct {
	Tag   DetailTag
	Value any
}

func (tv DetailTaggedValue) MarshalText() ([]byte, error) {
	return fmt.Appendf(nil, "[detail: %s]", SafeDetail), nil
}

func FilterDetails(d Details, tag DetailTag) Details {
	fd := Details{}
	for key, value := range d {
		if tv, ok := value.(DetailTaggedValue); ok && tv.Tag == tag {
			fd[key] = tv.Value
		}
	}
	return fd
}

func GetSafeDetails(err error) Details {
	d := CollectDetails(err, nil)
	return FilterDetails(d, SafeDetail)
}
