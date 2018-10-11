package json

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckKey(t *testing.T) {
	_, _, err := checkKey(nil, nil)
	assert.EqualError(t, err, ErrExpectedValue.Error())

	_, _, err = checkKey([]byte{'4'}, nil)
	assert.EqualError(t, err, ErrUnexpectedChar.Error())

	cases := []string{"", "key", `"""`, `\\\\`, `some' hard "ke\y`, "12.4"}
	for _, tc := range cases {
		data, err := json.Marshal(tc)
		if err != nil {
			t.Fatal(err)
		}

		// equal
		eq, off, err := checkKey(data, []byte(tc))
		assert.NoError(t, err, "error; for %v (%s)", tc, data)
		assert.Equal(t, len(data), off, "offset; for %v (%s)", tc, data)
		assert.True(t, eq, "equality; for %v (%s)", tc, data)

		// longer
		eq, off, err = checkKey(data, append([]byte(tc), '\\'))
		assert.NoError(t, err, "error; for %v (%s)", tc, data)
		assert.Equal(t, len(data), off, "offset; for %v (%s)", tc, data)
		assert.False(t, eq, "equality; for %v (%s)", tc, data)

		// shorter
		if tc == "" {
			continue
		}

		eq, off, err = checkKey(data, []byte(tc[:len(tc)/2]))
		assert.NoError(t, err, "error; for %v (%s)", tc, data)
		assert.Equal(t, len(data), off, "offset; for %v (%s)", tc, data)
		assert.False(t, eq, "equality; for %v (%s)", tc, data)
	}
}

func TestSkipString(t *testing.T) {
	_, err := skipString(nil)
	assert.EqualError(t, err, ErrExpectedValue.Error())

	_, err = skipString([]byte{'4'})
	assert.EqualError(t, err, ErrUnexpectedChar.Error())

	_, err = skipString([]byte(`"a`))
	assert.EqualError(t, err, ErrUnexpectedEnd.Error())

	cases := []string{"", "key", `"""`, `\\\\`, `some' hard "ke\y`, "12.4"}
	for _, tc := range cases {
		data, err := json.Marshal(tc)
		if err != nil {
			t.Fatal(err)
		}

		data = append(data, '\\')

		off, err := skipString(data)
		assert.NoError(t, err, "error; for %v (%s)", tc, data)
		assert.Equal(t, len(data)-1, off, "offset; for %v (%s)", tc, data)
	}
}

func TestSkipArray(t *testing.T) {
	_, err := skipArray(nil)
	assert.EqualError(t, err, ErrExpectedValue.Error())

	_, err = skipArray([]byte{'4'})
	assert.EqualError(t, err, ErrUnexpectedChar.Error())

	_, err = skipArray([]byte(`[[[]]`))
	assert.EqualError(t, err, ErrUnexpectedEnd.Error())

	_, err = skipArray([]byte(`[[][]]`))
	assert.EqualError(t, err, ErrUnexpectedChar.Error())

	cases := []string{"[]", `[[]]`, `[[[[]]]]`, `[[],[]]`, `[[],[[]],[[],[]]]`}
	for _, tc := range cases {
		data := []byte(tc)

		data = append(data, '\\')

		off, err := skipArray(data)
		assert.NoError(t, err, "error; for %v (%s)", tc, data)
		assert.Equal(t, len(data)-1, off, "offset; for %v (%s)", tc, data)
	}
}

func TestSkipObject(t *testing.T) {
	_, err := skipObject(nil)
	assert.EqualError(t, err, ErrExpectedValue.Error())

	_, err = skipObject([]byte{'4'})
	assert.EqualError(t, err, ErrUnexpectedChar.Error())

	_, err = skipObject([]byte(`{`))
	assert.EqualError(t, err, ErrUnexpectedEnd.Error())

	_, err = skipObject([]byte(`{"a"`))
	assert.EqualError(t, err, ErrUnexpectedEnd.Error())

	_, err = skipObject([]byte(`{"a"3`))
	assert.EqualError(t, err, ErrUnexpectedChar.Error())

	_, err = skipObject([]byte(`{"a":{}"a"}`))
	assert.EqualError(t, err, ErrUnexpectedChar.Error())

	_, err = skipObject([]byte(`{"a":{4}}`))
	assert.EqualError(t, err, ErrUnexpectedChar.Error())

	cases := []string{"{}", `{"key":{}}`, `{"a":{"b":{}}}`, `{"a":{},"b":{},"c":{}}`, `{"a":{"b":{},"c":{}}}`}
	for _, tc := range cases {
		data := []byte(tc)

		data = append(data, '\\')

		off, err := skipObject(data)
		assert.NoError(t, err, "error; for %s", data)
		assert.Equal(t, len(data)-1, off, "offset; for %s", data)
	}
}

func TestSkipBool(t *testing.T) {
	_, err := skipValue(nil)
	assert.EqualError(t, err, ErrExpectedValue.Error())

	cases := []string{`true`, `false`}
	for _, tc := range cases {
		data := []byte(tc)
		data = append(data, '\\')

		off, err := skipValue(data)
		assert.NoError(t, err, "error; for %s", data)
		assert.Equal(t, len(data)-1, off, "offset; for %s", data)
	}
}

func TestSkipNumber(t *testing.T) {
	off, err := skipValue([]byte(`12e`))
	assert.EqualError(t, err, ErrExpectedValue.Error())
	assert.Equal(t, 3, off, "12e")

	cases := []string{`0`, `1`, `123`, `-1`, `+1`, `1.1`, `1234567890`, `1e5`, `-1e+5`}
	for _, tc := range cases {
		data := []byte(tc)
		//	data = append(data, '\\')

		off, err = skipValue(data)
		assert.NoError(t, err, "error; for %s", data)
		assert.Equal(t, len(data), off, "offset; for %s", data)
	}
}

func TestSkipValueErr(t *testing.T) {
	cases := []string{`e`, `[}`, `{]`, `t`, `f`}
	for _, tc := range cases {
		data := []byte(tc)
		//	data = append(data, '\\')

		off, err := skipValue(data)
		assert.Error(t, err, "for %s", data)
		assert.Equal(t, len(data)-1, off, "offset; for %s", data)
	}
}

func TestSkipValueOK(t *testing.T) {
	cases := []string{`123`, `"abc"`, `[1,2]`, `[1,2,3,"abc",{"a":"b","c":[1,2,3]}]`}
	for _, tc := range cases {
		data := []byte(tc)

		off, err := skipValue(data)
		assert.NoError(t, err, "for %s", data)
		assert.Equal(t, len(data), off, "offset; for %s", data)

		if t.Failed() {
			break
		}
	}
}

func TestGetError(t *testing.T) {
	assert.Panics(t, func() {
		Wrap([]byte(``)).Get(struct{}{})
	}, struct{}{})

	cases := []struct {
		data string
		err  error
		keys []interface{}
	}{
		{data: `[0,1,2,3]`, err: ErrOutOfRange, keys: []interface{}{4}},
		{data: `[{"a"}]`, err: ErrUnexpectedChar, keys: []interface{}{0, "a"}},
	}

	for _, tc := range cases {
		j := Wrap([]byte(tc.data))
		_, err := j.Get(tc.keys...)
		assert.EqualError(t, err, tc.err.Error(), "for: '%s'", tc.data)
	}
}

func TestGetOK(t *testing.T) {
	cases := []struct {
		data string
		res  string
		keys []interface{}
	}{
		{data: `[]`, res: `[]`, keys: nil},
		{data: `{}`, res: `{}`, keys: nil},
		{data: `[1]`, res: `1`, keys: []interface{}{0}},
		{data: `[0,1,2,3,4]`, res: `2`, keys: []interface{}{2}},
		{data: `[0,1,2,3,4]`, res: `4`, keys: []interface{}{4}},
		{data: `{"a":1}`, res: `1`, keys: []interface{}{"a"}},
		{data: `{"a":1,"b":2,"c":3}`, res: `3`, keys: []interface{}{"c"}},
		{data: `{"a":1,"b":[0,1,2,3],"c":4}`, res: `2`, keys: []interface{}{[]byte("b"), 2}},
		{data: `{"a":1,"b":[0,1,true,3],"c":4}`, res: `true`, keys: []interface{}{"b", 2}},
	}

	for _, tc := range cases {
		j := Wrap([]byte(tc.data))
		res, err := j.Get(tc.keys...)
		assert.NoError(t, err, "for: '%s'", tc.data)
		assert.Equal(t, tc.res, string(res.Buffer()))
	}
}

func TestWrapUnsafe(t *testing.T) {
	v := WrapStringUnsafe(`[0,1,2]`)
	assert.Equal(t, 1, v.MustGet(1).MustInt())
}

func BenchmarkParseIntFmt(b *testing.B) {
	b.Skip()
	b.ReportAllocs()

	var v int
	buf := []byte("123456789012345")

	for i := 0; i < b.N; i++ {
		fmt.Sscanf("%d", string(buf), &v)
	}

	_ = v
}

func BenchmarkParseIntMy(b *testing.B) {
	b.ReportAllocs()

	var v int
	buf := []byte("123456789012345")
	j := Value{buf: buf}

	for i := 0; i < b.N; i++ {
		v, _ = j.Int()
	}

	_ = v
}
