package json

import (
	"encoding/hex"
	stdjson "encoding/json"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

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

func TestFuzzUnicode(t *testing.T) {
	const dir = "fuzz/FuzzUnicode_wd/corpus"

	fs, err := ioutil.ReadDir(dir)
	if !assert.NoError(t, err) {
		return
	}

	for _, f := range fs {
		if strings.Contains(f.Name(), ".") {
			//	t.Logf("skip file %v", f.Name())
			continue
		}

		d, err := ioutil.ReadFile(path.Join(dir, f.Name()))
		if !assert.NoError(t, err) {
			return
		}

		if !utf8.Valid(d) {
			panic("invalid string")
		}

		t.Run(f.Name(), func(t *testing.T) {
			if back, err := strconv.Unquote(strconv.Quote(string(d))); err != nil || back != string(d) {
				t.Logf("unquote * quote: %q -> %q  eq %v (%v)", d, back, back == string(d), err)
			} else {
				t.Logf("unquote * quote: %q -> %q  eq %v (%v)", d, back, back == string(d), err)
			}

			//	t.Logf("quoted:\n%v", hex.Dump([]byte(strconv.Quote(string(d)))))

			data, err := Marshal(string(d))
			assert.NoError(t, err, "marshal")

			var res string
			err = Unmarshal(data, &res)
			assert.NoError(t, err, "unmarshal")

			assert.Equal(t, d, []byte(res), "marshalled data:\n%v", hex.Dump(data))

			t.Logf("marsha:\n%v", hex.Dump(data))

			if t.Failed() {
				t.Logf("data len %d -> %d -> %d", len(d), len(data), len(res))

				stddata, stderr := stdjson.Marshal(string(d))
				if assert.NoError(t, stderr) {
					assert.Equal(t, stddata, data, "src data:\n%v", hex.Dump(d))
				}
			}
		})

		if t.Failed() {
			break
		}
	}
}
