package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {
	var r Reader

	for _, data := range []string{
		"null", "true", "false",
		"1", "1.1", "1e1", "+1", "-1", "-1.4", "0x1p+1", "-0x1p-2", "0x3", "0xf", "0XF",
		`""`, `"a"`, `"abc def"`, `"a\"b\nc\td"`,
		"[]", "[1, 2, 3]", `[null, "str"]`,
		"{}", `{"key":"val"}`, `{"k": "v", "k2": 3, "k3": [], "k4": {}, "k5": null}`,
	} {
		r.Reset([]byte(data), nil)

		raw, err := r.Raw()
		if !assert.NoError(t, err) || !assert.Equal(t, []byte(data), raw) {
			t.Logf("data: %q", data)
		}
	}
}
