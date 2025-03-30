package media

import (
	"fmt"
	"io"
	"net/http"
)

type Seeker struct {
	url    string
	client *http.Client
	offset int64
	size   int64
}

func (se *Seeker) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (se *Seeker) Seek(offset int64, whence int) (int64, error) {
	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = se.offset + offset
	case io.SeekEnd:
		newOffset = se.size + offset
	default:
		return 0, fmt.Errorf("invalid whence")
	}

	if newOffset < 0 {
		return 0, fmt.Errorf("negative position")
	}
	if newOffset > se.size {
		newOffset = se.size
	}
	se.offset = newOffset
	return newOffset, nil
}
