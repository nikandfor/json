package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	data, err := Marshal(nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte("null"), data)

	data, err = Marshal(1)
	assert.NoError(t, err)
	assert.Equal(t, []byte("1"), data)

	data, err = Marshal(1.1)
	assert.NoError(t, err)
	assert.Equal(t, []byte("1.1"), data)

	data, err = Marshal(true)
	assert.NoError(t, err)
	assert.Equal(t, []byte("true"), data)

	data, err = Marshal(false)
	assert.NoError(t, err)
	assert.Equal(t, []byte("false"), data)

	data, err = Marshal("string")
	assert.NoError(t, err)
	assert.Equal(t, []byte(`"string"`), data)
}

func TestMarshalSlice(t *testing.T) {
	data, err := Marshal([]byte("qwertbase64"))
	assert.NoError(t, err)
	assert.Equal(t, []byte(`"cXdlcnRiYXNlNjQ"`), data)

	data, err = Marshal([]int{1, 2, 3, 4, 5})
	assert.NoError(t, err)
	assert.Equal(t, []byte(`[1,2,3,4,5]`), data)

	data, err = Marshal([]string{"a", "b", "c"})
	assert.NoError(t, err)
	assert.Equal(t, []byte(`["a","b","c"]`), data)
}

func TestMarshalStruct(t *testing.T) {
	type B struct {
		E int    `json:"e"`
		D string `json:"d"`
	}
	type A struct {
		I  int     `json:"i"`
		Ip *int    `json:"ip"`
		S  string  `json:"s"`
		Sp *string `json:"sp"`
		B  []int   `json:"b"`
		Bp *[]int  `json:"bp"`
		A  B       `json:"a"`
		Ap *B      `json:"ap"`
	}

	data, err := Marshal(&B{E: 11, D: "d_str"})
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"e":11,"d":"d_str"}`), data)

	ival := 2
	sval := "sp_val"
	bval := []int{3, 2, 1}
	data, err = Marshal(&A{
		I:  1,
		Ip: &ival,
		S:  "s_val",
		Sp: &sval,
		B:  []int{1, 2, 3},
		Bp: &bval,
		A:  B{E: 4, D: "d_val"},
		Ap: nil,
	})
	assert.NoError(t, err)
	assert.Equal(t, `{"i":1,"ip":2,"s":"s_val","sp":"sp_val","b":[1,2,3],"bp":[3,2,1],"a":{"e":4,"d":"d_val"},"ap":null}`, string(data))
}
