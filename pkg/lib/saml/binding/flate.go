// Copied from https://github.com/crewjam/saml/blob/8e9236867d176ad6338c870a84e2039aef8a5021/flate.go

package binding

import (
	"compress/flate"
	"fmt"
	"io"
)

const flateUncompressLimit = 10 * 1024 * 1024 // 10MB

func NewSaferFlateReader(r io.Reader) io.ReadCloser {
	return &saferFlateReader{r: flate.NewReader(r)}
}

type saferFlateReader struct {
	r     io.ReadCloser
	count int
}

func (r *saferFlateReader) Read(p []byte) (n int, err error) {
	if r.count+len(p) > flateUncompressLimit {
		return 0, fmt.Errorf("flate: uncompress limit exceeded (%d bytes)", flateUncompressLimit)
	}
	n, err = r.r.Read(p)
	r.count += n
	return n, err
}

func (r *saferFlateReader) Close() error {
	return r.r.Close()
}
