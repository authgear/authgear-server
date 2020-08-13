package urlutil

import (
	"encoding/base64"
	"fmt"
	"io"
)

func DataURIWriter(mediatype string, w io.Writer) (out io.WriteCloser, err error) {
	_, err = fmt.Fprintf(w, "data:%s;base64,", mediatype)
	if err != nil {
		return
	}

	out = base64.NewEncoder(base64.StdEncoding, w)
	return
}
