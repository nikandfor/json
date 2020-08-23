package json

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuzzSkip(t *testing.T) {
	const dir = "fuzz/FuzzSkip_wd/corpus"

	fs, err := ioutil.ReadDir(dir)
	if !assert.NoError(t, err) {
		return
	}

	for _, f := range fs {
		data, err := ioutil.ReadFile(path.Join(dir, f.Name()))
		if !assert.NoError(t, err) {
			return
		}

		t.Run(f.Name(), func(t *testing.T) {
			Wrap(data).Skip()
		})
	}
}
