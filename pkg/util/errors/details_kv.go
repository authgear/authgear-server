package errors

import "fmt"

type DetailTag string

const (
	SafeDetail DetailTag = "safe"
)

func (t DetailTag) Value(value interface{}) DetailTaggedValue {
	return DetailTaggedValue{t, value}
}

type DetailTaggedValue struct {
	Tag   DetailTag
	Value interface{}
}

func (tv DetailTaggedValue) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("[detail: %s]", SafeDetail)), nil
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
