package function

import (
	"io"
	"io/ioutil"
)

func read(reader io.ReadCloser) []byte {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil
	}
	return b
}
