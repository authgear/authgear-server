package errors

type DetailTag string

const (
	SafeDetail DetailTag = "safe"
)

type DetailTaggedValue interface {
	IsTagged(tag DetailTag) bool
}

func FilterDetails(d Details, tags ...DetailTag) Details {
	fd := Details{}
	for key, value := range d {
		v, ok := value.(DetailTaggedValue)
		if !ok {
			continue
		}
		for _, tag := range tags {
			if v.IsTagged(tag) {
				fd[key] = value
				break
			}
		}
	}
	return fd
}

func GetSafeDetails(err error) Details {
	d := CollectDetails(err, nil)
	return FilterDetails(d, SafeDetail)
}

type SafeString string

func (SafeString) IsTagged(tag DetailTag) bool {
	return tag == SafeDetail
}
