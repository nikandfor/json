package json

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	var p Parser

	for _, d := range []string{
		"null", "true", "false",
		"1", "1.1", "1e1", "+1", "-1", "-1.4", "1p+1", "-1p-2", "0x3", "0xf", "0XF",
		`""`, `"a"`, `"abc def"`, `"a\"b\nc\td"`,
		"[]", "[1, 2, 3]", `[null, "str"]`,
		"{}", `{"key":"val"}`, `{"k": "v", "k2": 3, "k3": [], "k4": {}, "k5": null}`,
	} {
		i, err := p.Skip([]byte(d), 0)
		if !assert.NoError(t, err) || !assert.Equal(t, len(d), i) {
			t.Logf("pos: %d (%[1]x)  data: %q", i, d)
		}

		raw, i, err := p.Raw([]byte(d), 0)
		if !assert.NoError(t, err) || !assert.Equal(t, len(d), i) || !assert.Equal(t, raw, []byte(d)) {
			t.Logf("pos: %d (%[1]x)  data: %q", i, d)
		}
	}
}

func TestString(t *testing.T) {
	var p Parser

	for _, d := range []string{
		`""`, `"a"`, `"a\"b\nc\td"`,
	} {
		s, i, err := p.DecodeString([]byte(d), 0, nil)
		if !assert.NoError(t, err) || !assert.Equal(t, len(d), i) {
			t.Logf("pos: %d (%[1]x)  data: %q", i, d)
		}

		q, err := strconv.Unquote(d)
		assert.NoError(t, err)

		assert.Equal(t, q, string(s))
	}
}

func TestEnterMore(t *testing.T) {
	var p Parser

cases:
	for _, d := range []string{
		`{}`, `{ }`, `{"a":"b"}`, `{"a": "b", "c": 4}`,
		`[]`, `[ ]`, `["a"]`, `[2]`, `[ "a", 2, null ]`,
	} {
		b := []byte(d)

		i, err := p.Enter(b, 0, d[0])
		if !assert.NoError(t, err) {
			t.Logf("pos: %d (%[1]x)  data: %q", i, d)
			continue
		}

		for p.ForMore(b, &i, d[0], &err) {
			if d[0] == '{' {
				_, i, err = p.Key(b, i)
				if !assert.NoError(t, err) {
					continue cases
				}
			}

			i, err = p.Skip(b, i)
			if !assert.NoError(t, err) {
				continue cases
			}
		}

		if !assert.NoError(t, err) {
			t.Logf("pos: %d (%[1]x)  data: %q", i, d)
			continue
		}
	}
}
