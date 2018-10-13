package json

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckKey(t *testing.T) {
	_, s, err := checkKey(nil, nil, 0)
	assert.EqualError(t, err, ErrExpectedValue.Error())
	assert.Equal(t, 0, s)

	_, s, err = checkKey([]byte{'4'}, nil, 0)
	assert.EqualError(t, err, ErrUnexpectedChar.Error())
	assert.Equal(t, 0, s)

	cases := []string{"", "key", `"""`, `\\\\`, `some' hard "ke\y`, "12.4"}
	for _, tc := range cases {
		data, err := json.Marshal(tc)
		if err != nil {
			t.Fatal(err)
		}

		// equal
		eq, off, err := checkKey(data, []byte(tc), 0)
		assert.NoError(t, err, "error; for %v (%s)", tc, data)
		assert.Equal(t, len(data), off, "offset; for %v (%s)", tc, data)
		assert.True(t, eq, "equality; for %v (%s)", tc, data)

		// longer
		eq, off, err = checkKey(data, append([]byte(tc), '\\'), 0)
		assert.NoError(t, err, "error; for %v (%s)", tc, data)
		assert.Equal(t, len(data), off, "offset; for %v (%s)", tc, data)
		assert.False(t, eq, "equality; for %v (%s)", tc, data)

		// shorter
		if tc == "" {
			continue
		}

		eq, off, err = checkKey(data, []byte(tc[:len(tc)/2]), 0)
		assert.NoError(t, err, "error; for %v (%s)", tc, data)
		assert.Equal(t, len(data), off, "offset; for %v (%s)", tc, data)
		assert.False(t, eq, "equality; for %v (%s)", tc, data)
	}
}

func TestSkipString(t *testing.T) {
	//	i, err := skipString(nil, 0)
	//	assert.EqualError(t, err, ErrExpectedValue.Error())
	//	assert.Equal(t, 0, i)

	i, err := skipString([]byte{'4'}, 0)
	assert.EqualError(t, err, ErrUnexpectedChar.Error())
	assert.Equal(t, 0, i)

	i, err = skipString([]byte(`"a`), 0)
	assert.EqualError(t, err, ErrUnexpectedEnd.Error())
	assert.Equal(t, 2, i)

	cases := []string{"", "key", `"""`, `\\\\`, `some' hard "ke\y`, "12.4"}
	for _, tc := range cases {
		data, err := json.Marshal(tc)
		if err != nil {
			t.Fatal(err)
		}

		data = append(data, '\\')

		off, err := skipString(data, 0)
		assert.NoError(t, err, "error; for %v (%s)", tc, data)
		assert.Equal(t, len(data)-1, off, "offset; for %v (%s)", tc, data)
	}
}

func TestSkipArray(t *testing.T) {
	//	i, err := skipArray(nil, 0)
	//	assert.EqualError(t, err, ErrExpectedValue.Error())
	//	assert.Equal(t, 0, i)

	i, err := skipArray([]byte{'4'}, 0)
	assert.EqualError(t, err, ErrUnexpectedChar.Error())
	assert.Equal(t, 0, i)

	i, err = skipArray([]byte(`[[[]]`), 0)
	assert.EqualError(t, err, ErrUnexpectedEnd.Error())
	assert.Equal(t, 5, i)

	i, err = skipArray([]byte(`[[][]]`), 0)
	assert.EqualError(t, err, ErrUnexpectedChar.Error())
	assert.Equal(t, 3, i)

	cases := []string{"[]", `[[]]`, `[[[[]]]]`, `[[],[]]`, `[[],[[]],[[],[]]]`}
	for _, tc := range cases {
		data := []byte(tc)

		data = append(data, '\\')

		off, err := skipArray(data, 0)
		assert.NoError(t, err, "error; for '%s'", data)
		assert.Equal(t, len(data)-1, off, "offset; for '%s'", data)
	}
}

func TestSkipObject(t *testing.T) {
	//	i, err := skipObject(nil, 0)
	//	assert.EqualError(t, err, ErrExpectedValue.Error())
	//	assert.Equal(t, 0, i)

	i, err := skipObject([]byte{'4'}, 0)
	assert.EqualError(t, err, ErrUnexpectedChar.Error())
	assert.Equal(t, 0, i)

	i, err = skipObject([]byte(`{`), 0)
	assert.EqualError(t, err, ErrUnexpectedEnd.Error())
	assert.Equal(t, 1, i)

	i, err = skipObject([]byte(`{"a"`), 0)
	assert.EqualError(t, err, ErrUnexpectedEnd.Error())
	assert.Equal(t, 4, i)

	i, err = skipObject([]byte(`{"a"3`), 0)
	assert.EqualError(t, err, ErrUnexpectedChar.Error())
	assert.Equal(t, 4, i)

	i, err = skipObject([]byte(`{"a":{}"a"}`), 0)
	assert.EqualError(t, err, ErrUnexpectedChar.Error())
	assert.Equal(t, 7, i)

	i, err = skipObject([]byte(`{"a":{4}}`), 0)
	assert.EqualError(t, err, ErrUnexpectedChar.Error())
	assert.Equal(t, 6, i)

	cases := []string{"{}", `{"key":{}}`, `{"a":{"b":{}}}`, `{"a":{},"b":{},"c":{}}`, `{"a":{"b":{},"c":{}}}`}
	for _, tc := range cases {
		data := []byte(tc)

		data = append(data, '\\')

		off, err := skipObject(data, 0)
		assert.NoError(t, err, "error; for %s", data)
		assert.Equal(t, len(data)-1, off, "offset; for %s", data)
	}
}

func TestSkipLiteral(t *testing.T) {
	i, err := skipValue(nil, 0)
	assert.EqualError(t, err, ErrExpectedValue.Error())
	assert.Equal(t, 0, i)

	cases := []string{`true`, `false`, `null`}
	for _, tc := range cases {
		data := []byte(tc)
		data = append(data, '\\')

		off, err := skipValue(data, 0)
		assert.NoError(t, err, "error; for %s", data)
		assert.Equal(t, len(data)-1, off, "offset; for %s", data)
	}
}

func TestSkipNumber(t *testing.T) {
	off, err := skipValue([]byte(`12e`), 0) // exponent form
	assert.EqualError(t, err, ErrExpectedValue.Error())
	assert.Equal(t, 3, off, "12e")

	cases := []string{`0`, `1`, `123`, `-1`, `+1`, `1.1`, `1234567890`, `1e5`, `-1e+5`}
	for _, tc := range cases {
		data := []byte(tc)
		//	data = append(data, '\\')

		off, err = skipValue(data, 0)
		assert.NoError(t, err, "error; for %s", data)
		assert.Equal(t, len(data), off, "offset; for %s", data)
	}

	for _, tc := range cases {
		data := []byte(tc)
		data = append(data, '\\')

		off, err = skipValue(data, 0)
		assert.NoError(t, err, "error; for %s", data)
		assert.Equal(t, len(data)-1, off, "offset; for %s", data)
	}
}

func TestSkipValueErr(t *testing.T) {
	cases := []string{`e`, `[}`, `{]`, `t`, `f`}
	for _, tc := range cases {
		data := []byte(tc)

		off, err := skipValue(data, 0)
		assert.Error(t, err, "for %s", data)
		assert.Equal(t, len(data)-1, off, "offset; for %s", data)
	}
}

func TestSkipValueOK(t *testing.T) {
	cases := []string{`123`, `"abc"`, `[1,2]`, `[1,2,3,"abc",{"a":"b","c":[1,2,3]}]`}
	for _, tc := range cases {
		data := []byte(tc)

		off, err := skipValue(data, 0)
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
		assert.EqualError(t, err, tc.err.Error(), "for: '%s' get %v", tc.data, tc.keys)
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
		{data: ` { "a" : 1 , "b" : [ 1 , false , { } , null ], "c" : [ 0 , 1 , true , 3 ] , "d" : 4 } `, res: `null`, keys: []interface{}{"b", 3}},
	}

	for _, tc := range cases {
		j := Wrap([]byte(tc.data))
		res, err := j.Get(tc.keys...)
		if assert.NoError(t, err, "for: '%s' keys %v", tc.data, tc.keys) {
			assert.Equal(t, tc.res, string(res.Buffer()), "for '%s' keys %v", tc.data, tc.keys)
		}
	}
}

func TestSkipSpace(t *testing.T) {
	i := skipSpaces(nil, 0)
	assert.Equal(t, 0, i)

	cases := []string{" ", "      ", "  \n  ", "\t\t\n\t\t\n"}
	for _, tc := range cases {
		i := skipSpaces([]byte(tc), 0)
		assert.Equal(t, len(tc), i)
	}

	for _, tc := range cases {
		i := skipSpaces(append([]byte(tc), '1'), 0)
		assert.Equal(t, len(tc), i)
	}
}

func TestString(t *testing.T) {
	cases := []string{`123`, `[1,2]`, `[1,2,3,"abc",{"a":"b","c":[1,2,3]}]`}
	for _, tc := range cases {
		v := WrapString(tc)
		assert.Equal(t, tc, v.String())
	}

	cases = []string{`"abc"`, `"123"`, `"[]"`}
	for _, tc := range cases {
		v := WrapString(tc)
		assert.Equal(t, tc[1:len(tc)-1], v.String())
	}

	r := WrapString(`{"a":["a","b","c"]}`).MustGet("a", 2).String()
	assert.Equal(t, "c", r)
}

func TestErrorExtended(t *testing.T) {
	var cases = []struct {
		data    string
		pos     int
		skip    int
		err     string
		skipadd bool
	}{
		{pos: 0, skip: 0, data: `[1,2,3]`, err: "<nil>", skipadd: true},
		{pos: -2, skip: -2, data: `[1,2,3 3]`, err: "parse error at pos 7: unexpected char\n[1,2,3 3]\n_______^_\n"},
		{pos: 0, skip: 0, data: `tru`, err: "parse error at pos 0: unexpected char\ntru\n^__\n"},
		{pos: -1, skip: -1, data: `[1,2,3}`, err: "parse error at pos 6: unexpected char\n[1,2,3}\n______^\n"},
		{pos: 3, skip: 3, data: `[1,q,3]`, err: "parse error at pos 3: expected value\n[1,q,3]\n___^___\n"},
		{pos: -1, skip: -1, data: `[1,2,3q`, err: "parse error at pos 6: unexpected char\n[1,2,3q\n______^\n"},
		{pos: 1, skip: 1, data: `[q]`, err: "parse error at pos 1: expected value\n[q]\n_^_\n"},
		{pos: 1, skip: 1, data: `{q}`, err: "parse error at pos 1: unexpected char\n{q}\n_^_\n"},
		{pos: 4, skip: 4, data: `{"a"}`, err: "parse error at pos 4: unexpected char\n{\"a\"}\n____^\n"},
		{pos: 5, skip: 5, data: `{"a":}`, err: "parse error at pos 5: expected value\n{\"a\":}\n_____^\n"},
		{pos: 5, skip: 5, data: `{"a":qwe}`, err: "parse error at pos 5: expected value\n{\"a\":qwe}\n_____^___\n"},
		{pos: 2, skip: 2, data: `[{q}]`, err: "parse error at pos 2: unexpected char\n[{q}]\n__^__\n"},
		{pos: 6, skip: 6, data: `{"a":[q]}`, err: "parse error at pos 6: expected value\n{\"a\":[q]}\n______^__\n"},
		{pos: -2, skip: -2, data: `{"first":[1,2,"three",[4]],"second":{"a":1,"b":2}q}`, err: `parse error at pos 49: unexpected char
{"first":[1,2,"three",[4]],"second":{"a":1,"b":2}q}
_________________________________________________^_
`},
		{pos: 92, skip: 92, data: `   {
			"first" : [  
				1 , 2	,	"three" , [ 4 ] ] , 
			"second" : { "a" : 1,
						"b":2}q}  
				`,
			err: `parse error at pos 92: unexpected char
... 4 ] ] , \n\t\t\t"second" : { "a" : 1,\n\t\t\t\t\t\t"b":2}q}  \n\t\t\t\t
_____________________________________________________________^_____________
`},
	}

	for id, tc := range cases {
		if tc.skipadd || tc.skip < 0 {
			tc.skip += len(tc.data)
		}
		if tc.pos < 0 {
			tc.pos += len(tc.data)
		}

		i, err := skipValue([]byte(tc.data), 0)
		assert.Equal(t, tc.skip, i, "skip for test case: %v", id)
		if e, ok := err.(Error); ok {
			e.b = []byte(tc.data)
			err = e
			assert.Equal(t, tc.pos, e.Pos(), "pos for test case: %v", id)
		}
		aerr := fmt.Sprintf("%+v", err)
		assert.Equal(t, tc.err, aerr, "error message for test case: %v", id)
		if t.Failed() {
			t.Logf("expected\n%v", tc.err)
			t.Logf("actual\n%v", aerr)
			break
		}
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
	j := Wrap(buf)

	for i := 0; i < b.N; i++ {
		v, _ = j.Int()
	}

	_ = v
}
