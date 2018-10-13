package json

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

const dir = "fuzz/wd/corpus"

func TestFuzz(t *testing.T) {
	t.Skip("it must just not crash")

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
			var p interface{}
			e1 := json.Unmarshal(data, &p)
			_, err = Parse(data)
			if e1 == nil {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err, "stdlib: %v\n%q", e1, data)
		})
	}
}

func TestDecodeString(t *testing.T) {
	t.Skip("it shouldn't pass")

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
			data = append([]byte{'"'}, data...)
			data = append(data, '"')
			_, err := decodeString(data, 0)
			assert.NoError(t, err)
		})
	}
}
