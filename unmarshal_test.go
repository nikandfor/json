package json

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReflect(tb *testing.T) {
	var x int
	var xp *int

	f := func(x interface{}) {
		r := reflect.TypeOf(x)
		tp, d := unpack(x)

		tb.Logf("type %T: ptr %5v  indir %5v  tp/d %x %x", x, r.Kind() == reflect.Pointer, ifaceIndir(tp), tp, d)
	}

	f(x)
	f(xp)
	f(&xp)
}

func TestUnmarshal(tb *testing.T) {
	type (
		O1 struct {
			N int    `json:"n"`
			S string `json:"s"`

			N2 *int    `json:"n2"`
			S2 *string `json:"s2"`
		}
	)

	var (
		d Decoder

		x   int
		x64 int64

		xptr  = &x
		xptr2 = &xptr

		s string

		o1 O1
	)

	_, err := d.Unmarshal([]byte(`3`), 0, &x)
	assert.NoError(tb, err)
	assert.Equal(tb, 3, x)

	_, err = d.Unmarshal([]byte(`5`), 0, &x64)
	assert.NoError(tb, err)
	assert.Equal(tb, int64(5), x64)

	_, err = d.Unmarshal([]byte(`4`), 0, &xptr)
	assert.NoError(tb, err)
	assert.Equal(tb, 4, x)

	_, err = d.Unmarshal([]byte(`5`), 0, &xptr2)
	assert.NoError(tb, err)
	assert.Equal(tb, 5, x)

	_, err = d.Unmarshal([]byte(`"abc"`), 0, &s)
	assert.NoError(tb, err)
	assert.Equal(tb, "abc", s)

	_, err = d.Unmarshal([]byte(`true`), 0, &x)
	assert.ErrorIs(tb, err, ErrType)

	_, err = d.Unmarshal([]byte(`{"n":4,"s":"abc","n2":6,"s2":"qwe"}`), 0, &o1)
	assert.NoError(tb, err)
	assert.Equal(tb, O1{
		N: 4,
		S: "abc",
		N2: func() *int {
			x := 6
			return &x
		}(),
		S2: func() *string {
			x := "qwe"
			return &x
		}(),
	}, o1)
}
