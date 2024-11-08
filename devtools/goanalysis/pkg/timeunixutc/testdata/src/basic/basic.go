package basic

import (
	"time"
)

func Example() {
	_ = time.Unix(0, 0).UTC()
	_ = time.UnixMilli(0).UTC()
	_ = time.UnixMicro(0).UTC()

	_ = time.Unix(0, 0)   // want "time.Unix\\(\\) is not immediately followed by .UTC\\(\\)"
	_ = time.UnixMilli(0) // want "time.UnixMilli\\(\\) is not immediately followed by .UTC\\(\\)"
	_ = time.UnixMicro(0) // want "time.UnixMicro\\(\\) is not immediately followed by .UTC\\(\\)"

	_ = time.Unix(0, 0).Add(0).UTC()   // want "time.Unix\\(\\) is not immediately followed by .UTC\\(\\)"
	_ = time.UnixMilli(0).Add(0).UTC() // want "time.UnixMilli\\(\\) is not immediately followed by .UTC\\(\\)"
	_ = time.UnixMicro(0).Add(0).UTC() // want "time.UnixMicro\\(\\) is not immediately followed by .UTC\\(\\)"
}
