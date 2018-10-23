package json

import (
	"encoding/base64"
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

func TestUnmarshalSliceOfInts(t *testing.T) {
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

func TestUnmarshalStringIntoSliceOfBytes(t *testing.T) {
	a := []byte{3, 2}
	err := WrapString("null").Unmarshal(&a)
	if assert.NoError(t, err) {
		assert.Equal(t, []byte(nil), a)
	}

	a = []byte(nil)
	err = WrapString(`""`).Unmarshal(&a)
	if assert.NoError(t, err) {
		assert.Equal(t, []byte{}, a)
	}

	a = []byte{3, 2}
	err = WrapString(`""`).Unmarshal(&a)
	if assert.NoError(t, err) {
		assert.Equal(t, []byte{}, a)
	}

	a = []byte{3, 2}
	err = WrapString(`"` + base64.StdEncoding.EncodeToString([]byte("qwe")) + `"`).Unmarshal(&a)
	if assert.NoError(t, err) {
		assert.Equal(t, []byte("qwe"), a)
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

	b := B{E: 3, D: "d_not_empty"}
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

func TestUnmarshalNoZero(t *testing.T) {
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

	sval = "sp_2"
	err = UnmarshalNoZero([]byte(`{"i":10,"sp":"sp_2"}`), &a)
	if assert.NoError(t, err) {
		assert.Equal(t, A{
			I:  10,
			Ip: &ival,
			S:  "s_val",
			Sp: &sval,
			B:  []byte{1, 2, 3},
			Bp: &bval,
			A:  B{E: 4, D: "d_val"},
			Ap: nil,
		}, a)
	}
}

func TestUnmarshalMap(t *testing.T) {
	data := `{"i":1,"ip":2,"s":"s_val","sp":"sp_val","b":[1,2,3],"bp":[3,2,1],"a":{"e":4,"d":"d_val"},"ap":null}`
	var m map[string]interface{}
	err := WrapString(data).Unmarshal(&m)
	assert.NoError(t, err)
	exp := map[string]interface{}{
		"i":  Num("1"),
		"ip": Num("2"),
		"s":  "s_val",
		"sp": "sp_val",
		"b":  []interface{}{Num("1"), Num("2"), Num("3")},
		"bp": []interface{}{Num("3"), Num("2"), Num("1")},
		"a": map[string]interface{}{
			"e": Num("4"),
			"d": "d_val",
		},
		"ap": nil,
	}
	if !assert.Equal(t, exp, m) {
		for k, v := range exp {
			t.Logf("%10s : %v vs %v", k, v, m[k])
		}
		return
	}

	var i interface{}
	err = WrapString(data).Unmarshal(&i)
	assert.NoError(t, err)
	assert.Equal(t, exp, i)
}

func TestUnmarshalMapString(t *testing.T) {
	data := `{"a": "b", "c": "d", "e": "f"}`
	var m map[string]string
	err := WrapString(data).Unmarshal(&m)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"a": "b",
		"c": "d",
		"e": "f",
	}, m)
}

func TestUnmarshalMapInt(t *testing.T) {
	data := `{"a": 1, "c": 2, "e": 3}`
	var m map[string]int
	err := WrapString(data).Unmarshal(&m)
	assert.NoError(t, err)
	assert.Equal(t, map[string]int{
		"a": 1,
		"c": 2,
		"e": 3,
	}, m)
}

func TestUnmarshalMapFloat(t *testing.T) {
	data := `{"a": 1, "c": 2.3, "e": 3.333}`
	var m map[string]float64
	err := WrapString(data).Unmarshal(&m)
	assert.NoError(t, err)
	assert.Equal(t, map[string]float64{
		"a": 1,
		"c": 2.3,
		"e": 3.333,
	}, m)
}
