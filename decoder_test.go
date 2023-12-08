package json

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecoder(t *testing.T) {
	var d Decoder

	for _, data := range []string{
		"null", "true", "false",
		"1", "1.1", "1e1", "+1", "-1", "-1.4", "0x1p+1", "-0x1p-2", "0x3", "0xf", "0XF",
		`""`, `"a"`, `"abc def"`, `"a\"b\nc\td"`,
		"[]", "[1, 2, 3]", `[null, "str"]`,
		"{}", `{"key":"val"}`, `{"k": "v", "k2": 3, "k3": [], "k4": {}, "k5": null}`,
	} {
		raw, i, err := d.Raw([]byte(data), 0)
		if !assert.NoError(t, err) || !assert.Equal(t, len(data), i) || !assert.Equal(t, raw, []byte(data)) {
			t.Logf("pos: %d (%[1]x)  data: %q", i, data)
		}
	}
}

func TestString(t *testing.T) {
	var d Decoder

	for j, data := range []string{
		`""`, `"a"`, `"a\"b\nc\tde\"f\\g"`,
		//	`"\xab\xac\xf3"`,
		`"\u00ab\u00ac\u00f3"`,
		`"\u0100\u017e"`,
		//	`"\U00e4b896\U00e7958c"`,
	} {
		i, err := d.Skip([]byte(data), 0)
		if !assert.NoError(t, err) || !assert.Equal(t, len(data), i) {
			t.Logf("pos: %d (%[1]x)  data: %d %q", i, j, data)
			continue
		}

		s, i, err := d.DecodeString([]byte(data), 0, nil)
		if !assert.NoError(t, err) || !assert.Equal(t, len(data), i) {
			t.Logf("pos: %d (%[1]x)  data: %d %q", i, j, data)
			continue
		}

		var q string

		err = json.Unmarshal([]byte(data), &q)
		assert.NoError(t, err)
		assert.Equal(t, q, string(s))

		if false {
			q, err := strconv.Unquote(data)
			assert.NoError(t, err)

			assert.Equal(t, q, string(s))
		}
	}
}

func TestEnterMore(t *testing.T) {
	var d Decoder

cases:
	for _, data := range []string{
		`{}`, `{ }`, `{"a":"b"}`, `{"a": "b", "c": 4}`,
		`[]`, `[ ]`, `["a"]`, `[2]`, `[ "a", 2, null ]`,
	} {
		b := []byte(data)

		i, err := d.Enter(b, 0, data[0])
		if !assert.NoError(t, err) {
			t.Logf("pos: %d (%[1]x)  data: %q", i, data)
			continue
		}

		for d.ForMore(b, &i, data[0], &err) {
			if data[0] == '{' {
				_, i, err = d.Key(b, i)
				if !assert.NoError(t, err) {
					continue cases
				}
			}

			i, err = d.Skip(b, i)
			if !assert.NoError(t, err) {
				continue cases
			}
		}

		if !assert.NoError(t, err) {
			t.Logf("pos: %d (%[1]x)  data: %q", i, data)
			continue
		}
	}
}

func BenchmarkIsDigit(b *testing.B) {
	b.Run("RangeCmp", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			isDigit1('f', true)
		}
	})

	b.Run("BitOps", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			isDigit1('f', true)
		}
	})
}
