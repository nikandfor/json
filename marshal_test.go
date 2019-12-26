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
	assert.Equal(t, []byte(`"cXdlcnRiYXNlNjQ="`), data)

	data, err = Marshal([]int{1, 2, 3, 4, 5})
	assert.NoError(t, err)
	assert.Equal(t, []byte(`[1,2,3,4,5]`), data)

	data, err = Marshal([]string{"a", "b", "c"})
	assert.NoError(t, err)
	assert.Equal(t, []byte(`["a","b","c"]`), data)
}

func TestMarshalStruct(t *testing.T) {
	data, err := Marshal(&B{E: 11, D: "d_str"})
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"e":11,"d":"d_str"}`), data)

	ival := 2
	sval := "sp_val"
	bval := []int{3, 2, 1}
	data, err = Marshal(&A{
		I:   1,
		Ip:  &ival,
		S:   "s_val",
		Sp:  &sval,
		Is:  []int{1, 2, 3},
		Isp: &bval,
		A:   B{E: 4, D: "d_val"},
		Ap:  nil,
	})
	assert.NoError(t, err)
	assert.Equal(t, `{"i":1,"ip":2,"s":"s_val","sp":"sp_val","is":[1,2,3],"isp":[3,2,1],"a":{"e":4,"d":"d_val"},"ap":null}`, string(data))
}

func TestMarshalMap(t *testing.T) {
	data, err := Marshal(map[string]interface{}{
		"e": 11, "d": "d_str",
	})
	assert.NoError(t, err)
	assert.Subset(t, [][]byte{
		[]byte(`{"e":11,"d":"d_str"}`),
		[]byte(`{"d":"d_str","e":11}`),
	}, [][]byte{data}, "%s", data)

	data, err = Marshal(map[string]string{
		"e": "11", "d": "d_str",
	})
	assert.NoError(t, err)
	assert.Subset(t, [][]byte{
		[]byte(`{"e":"11","d":"d_str"}`),
		[]byte(`{"d":"d_str","e":"11"}`),
	}, [][]byte{data}, "%s", data)

	data, err = Marshal(map[string]int64{
		"e": 11, "d": 44,
	})
	assert.NoError(t, err)
	assert.Subset(t, [][]byte{
		[]byte(`{"e":11,"d":44}`),
		[]byte(`{"d":44,"e":11}`),
	}, [][]byte{data}, "%s", data)
}

func TestMapStructPtr(t *testing.T) {
	data, err := Marshal(map[string]interface{}{
		"b": &B{E: 12, D: "d_str"},
	})
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"b":{"e":12,"d":"d_str"}}`), data)
}

type B struct {
	E int     `json:"e"`
	D string  `json:"d"`
	U uintptr `json:"u,omitempty"`
}
type A struct {
	I   int                         `json:"i"`
	Ip  *int                        `json:"ip"`
	S   string                      `json:"s"`
	Sp  *string                     `json:"sp"`
	B   []byte                      `json:"b,omitempty"`
	Bp  *[]byte                     `json:"bp,omitempty"`
	Is  []int                       `json:"is"`
	Isp *[]int                      `json:"isp"`
	A   B                           `json:"a"`
	Ap  *B                          `json:"ap"`
	M   map[interface{}]interface{} `json:"m,omitempty"`
}

var (
	n  = 400
	s  = "pointer_to_str"
	b  = []byte("byte_slice_ptr")
	is = []int{7, 6, 5, 4, 3, 2, 1, 0}
	v  = &A{
		I:   4,
		Ip:  &n,
		S:   "string",
		Sp:  &s,
		B:   []byte("byte_slice"),
		Bp:  &b,
		Is:  []int{0, 1, 2, 3, 4, 5, 6, 7},
		Isp: &is,
		A:   B{E: 66, D: "b_str", U: 90000},
		Ap:  &B{E: 77, D: "b_ptr_str", U: 30000},
		M: map[interface{}]interface{}{
			"string": "string",
			"int":    99933,
			"B":      B{E: 44, D: "map_b_str"},
			"Bp":     &B{E: 33, D: "map_b_ptr_str"},
			"bytes":  []byte("map_bytes"),
		},
	}
)

func TestMarshalBigValue(t *testing.T) {
	data, err := Marshal(v)
	assert.NoError(t, err)

	var r A
	err = Unmarshal(data, &r)
	assert.NoError(t, err)

	exp := &(*v)
	exp.M = map[interface{}]interface{}{
		"string": "string",
		"int":    Num([]byte("99933")),
		"B": map[string]interface{}{
			"e": Num([]byte("44")),
			"d": "map_b_str",
		},
		"Bp": map[string]interface{}{
			"e": Num([]byte("33")),
			"d": "map_b_ptr_str",
		},
		"bytes": "bWFwX2J5dGVz",
	}

	assert.Equal(t, exp, &r, "%s", data)
}

func BenchmarkMarshal(b *testing.B) {
	b.ReportAllocs()

	var err1, err2 error
	var data []byte
	var r A
	for i := 0; i < b.N; i++ {
		data, err1 = Marshal(v)
		err2 = Unmarshal(data, &r)
	}
	assert.NoError(b, err1)
	assert.NoError(b, err2)
}
