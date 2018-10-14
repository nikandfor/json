package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSkipString(t *testing.T) {
	data := `"just_strings""next_doc"`
	it := WrapString(data)

	it.Skip()

	rd := it.BytesRead()
	tp := it.TypeNext()
	err := it.Err()

	t.Logf("tp %d %v %v", rd, tp, err)

	assert.NoError(t, err)
	assert.Equal(t, String, tp)
	assert.Equal(t, rd, 14)
}

func TestSkipNumbers(t *testing.T) {
	data := `1234 4321`
	it := WrapString(data)

	it.Skip()

	rd := it.BytesRead()
	tp := it.TypeNext()
	err := it.Err()

	t.Logf("tp %d %v %v", rd, tp, err)

	assert.NoError(t, err)
	assert.Equal(t, Number, tp)
	assert.Equal(t, rd, 4)
}

func TestSkipSparse(t *testing.T) {
	data := ` { "a" : "b" , "c" : [ "d" , 2 , "f" , null , [ true , [ ] , { } ] ] , "g" : false } `
	it := WrapString(data)

	it.Skip()

	rd := it.BytesRead()
	tp := it.TypeNext()
	err := it.Err()

	t.Logf("tp %d %v %v", rd, tp, err)

	assert.NoError(t, err)
	assert.Equal(t, None, tp)
	assert.Equal(t, rd, len(data)-1)
}

func TestSkipDense(t *testing.T) {
	data := `{"a":"b","c":["d",2,"f",null,[true,[],{}]],"g":false}`
	it := WrapString(data)

	it.Skip()

	rd := it.BytesRead()
	tp := it.TypeNext()
	err := it.Err()

	t.Logf("tp %d %v %v", rd, tp, err)

	assert.NoError(t, err)
	assert.Equal(t, None, tp)
	assert.Equal(t, rd, len(data))
}

func TestGetDense(t *testing.T) {
	data := `{"a":"b","c":["d",2,"f",null,[true,[],{}]],"g":false}`
	it := WrapString(data)

	it.Get("c", 4, 0)
	assert.NoError(t, it.err)
	assert.Equal(t, 30, it.BytesRead())
	assert.Equal(t, Bool, it.TypeNext())

	b, err := it.Bool()

	assert.NoError(t, err)
	assert.Equal(t, true, b)
	assert.Equal(t, 34, it.BytesRead())
	assert.Equal(t, ArrayStart, it.TypeNext())
}
