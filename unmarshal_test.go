package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalLiteral(t *testing.T) {
	var i int
	err := WrapString("4").Unmarshal(&i)
	if assert.NoError(t, err) {
		assert.Equal(t, 4, i)
	}

	var s string
	err = WrapString(`"str_val"`).Unmarshal(&s)
	if assert.NoError(t, err) {
		assert.Equal(t, "str_val", s)
	}
}

func TestUnmarshalLiteralPtr(t *testing.T) {
	var i *int
	err := WrapString("4").Unmarshal(&i)
	if assert.NoError(t, err) {
		assert.Equal(t, 4, *i)
	}

	var s *string
	err = WrapString(`"str_val"`).Unmarshal(&s)
	if assert.NoError(t, err) {
		assert.Equal(t, "str_val", *s)
	}
}

func TestUnmarshalSlice(t *testing.T) {
	a := []int{3, 2}

	err := WrapString("null").Unmarshal(&a)
	if assert.NoError(t, err) {
		assert.Equal(t, []int(nil), a)
	}

	a = []int(nil)
	err = WrapString("[]").Unmarshal(&a)
	if assert.NoError(t, err) {
		assert.Equal(t, []int{}, a)
	}

	a = []int{3, 2}
	err = WrapString("[]").Unmarshal(&a)
	if assert.NoError(t, err) {
		assert.Equal(t, []int{}, a)
	}

	err = WrapString("[1,2,3]").Unmarshal(&a)
	if assert.NoError(t, err) {
		assert.Equal(t, []int{1, 2, 3}, a)
	}
}

func TestUnmarshalSliceOfPtr(t *testing.T) {
	a := []*int{new(int)}

	err := WrapString("null").Unmarshal(&a)
	if assert.NoError(t, err) {
		assert.Equal(t, []*int(nil), a)
	}

	err = WrapString("[]").Unmarshal(&a)
	if assert.NoError(t, err) {
		assert.Equal(t, []*int{}, a)
	}

	vals := []int{1, 2, 3}
	err = WrapString("[1,2,3]").Unmarshal(&a)
	if assert.NoError(t, err) {
		assert.Equal(t, []*int{&vals[0], &vals[1], &vals[2]}, a)
	}

	vals = []int{4, 5, 6}
	ptrs := []int{0, 0, 0}
	a = []*int{&ptrs[0], &ptrs[1], &ptrs[2]}
	err = WrapString("[4,5,6]").Unmarshal(&a)
	if assert.NoError(t, err) {
		assert.Equal(t, []*int{&vals[0], &vals[1], &vals[2]}, a)
		assert.Equal(t, vals, ptrs)
	}
}

func TestUnmarshalArray(t *testing.T) {
	type A [3]int
	a := A{3, 2, 1}

	err := WrapString("[]").Unmarshal(&a)
	if assert.NoError(t, err) {
		assert.Equal(t, A{}, a)
	}

	err = WrapString("[1,2,3]").Unmarshal(&a)
	if assert.NoError(t, err) {
		assert.Equal(t, A{1, 2, 3}, a)
	}
}

func TestUnmarshalStruct(t *testing.T) {
	type B struct {
		E int
		D string
	}

	var b B
	err := WrapString(`{}`).Unmarshal(&b)
	if assert.NoError(t, err) {
		assert.Equal(t, B{}, b)
	}

	err = WrapString(`{"e": 5, "d": "d_val"}`).Unmarshal(&b)
	if assert.NoError(t, err) {
		assert.Equal(t, B{E: 5, D: "d_val"}, b)
	}
}

func TestUnmarshalStructNested(t *testing.T) {
	type B struct {
		E int
		D string
	}
	type A struct {
		I  int
		Ip *int
		S  string
		Sp *string
		B  []byte
		Bp *[]byte
		A  B
		Ap *B
	}

	var a A

	ival := 2
	sval := "sp_val"
	bval := []byte{3, 2, 1}
	err := WrapString(`{"i":1,"ip":2,"s":"s_val","sp":"sp_val","b":[1,2,3],"bp":[3,2,1],"a":{"e":4,"d":"d_val"},"ap":null}`).Unmarshal(&a)
	if assert.NoError(t, err) {
		assert.Equal(t, A{
			I:  1,
			Ip: &ival,
			S:  "s_val",
			Sp: &sval,
			B:  []byte{1, 2, 3},
			Bp: &bval,
			A:  B{E: 4, D: "d_val"},
			Ap: nil,
		}, a)
	}

	err = WrapString(`{}`).Unmarshal(&a)
	if assert.NoError(t, err) {
		assert.Equal(t, A{}, a)
	}
}

func TestUnmarshalStructNestedPtr(t *testing.T) {
	type B struct {
		E int
		D string
	}
	type A struct {
		I  int
		Ip *int
		S  string
		Sp *string
		B  []byte
		Bp *[]byte
		A  B
		Ap *B
	}

	var a A

	ival := 2
	sval := "sp_val"
	bval := []byte{3, 2, 1}
	err := WrapString(`{"i":1,"ip":2,"s":"s_val","sp":"sp_val","b":[1,2,3],"bp":[3,2,1],"a":{"e":4,"d":"d_val"},"ap":{"e":6,"d":"ptr_d_val"}}`).Unmarshal(&a)
	if assert.NoError(t, err) {
		assert.Equal(t, A{
			I:  1,
			Ip: &ival,
			S:  "s_val",
			Sp: &sval,
			B:  []byte{1, 2, 3},
			Bp: &bval,
			A:  B{E: 4, D: "d_val"},
			Ap: &B{E: 6, D: "ptr_d_val"},
		}, a)
	}

	var b = a

	err = WrapString(`{"i":1,"ip":2,"s":"s_val","sp":"sp_val","b":[1,2,3],"bp":[3,2,1],"a":{"e":5,"d":"d_val2"},"ap":{"e":7,"d":"ptr_d_val2"}}`).Unmarshal(&b)
	if assert.NoError(t, err) {
		assert.Equal(t, A{
			I:  1,
			Ip: &ival,
			S:  "s_val",
			Sp: &sval,
			B:  []byte{1, 2, 3},
			Bp: &bval,
			A:  B{E: 4, D: "d_val"},
			Ap: &B{E: 7, D: "ptr_d_val2"},
		}, a)
	}
}
