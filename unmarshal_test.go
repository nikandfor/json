package json

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestUnsafe(tb *testing.T) {
	type Rec struct {
		X int    `json:"x"`
		S string `json:"s"`

		Next *Rec `json:"next"`
	}

	tp, _ := unpack(Rec{})

	tb.Logf("type %v %x", tpString(tp), tp)

	p0 := unsafe_New(tp)
	p1 := unsafe_New(tp)
	p2 := unsafe_New(tp)

	tb.Logf("new %v: %p %p %p", tpString(tp), p0, p1, p2)
}

func TestUnmarshal(tb *testing.T) {
	type (
		Int struct {
			N int `json:"n"`
		}

		IntPtr struct {
			N *int `json:"n"`
		}

		Str struct {
			S string `json:"s"`
		}

		StrPtr struct {
			S *string `json:"s"`
		}

		IntStr struct {
			N int    `json:"n"`
			S string `json:"s"`
		}

		IntStrPtr struct {
			N *int    `json:"n"`
			S *string `json:"s"`
		}

		Rec struct {
			X int    `json:"x"`
			S string `json:"s"`

			Next *Rec `json:"next"`
		}
	)

	for _, tc := range []struct {
		N string
		D string
		X any
		E any
	}{
		{"int", `3`, new(int), 3},
		{"int64", `-4`, new(int64), int64(-4)},
		{"*int", `5`, ptr(new(int)), ptr(5)},

		{"string", `"abc"`, new(string), "abc"},
		{"*string", `"qwe"`, ptr(new(string)), ptr("qwe")},

		{"Int", `{"n":6}`, new(Int), Int{N: 6}},
		{"IntPtr", `{"n":7}`, new(IntPtr), IntPtr{N: ptr(7)}},

		{"Str", `{"s":"abc"}`, new(Str), Str{S: "abc"}},
		{"StrPtr", `{"s":"qwe"}`, new(StrPtr), StrPtr{S: ptr("qwe")}},

		{"IntStr", `{"n":8,"s":"abc"}`, new(IntStr), IntStr{N: 8, S: "abc"}},
		{"IntStrPtr", `{"n":9,"s":"qwe"}`, new(IntStrPtr), IntStrPtr{N: ptr(9), S: ptr("qwe")}},

		{"Rec", `
{
	"x": 1,
	"s": "one",
	"next": {
		"x": 2,
		"next": {
			"next": null,
			"x": 3,
			"s": "three"
		},
		"s":"two"
	}
}`, new(Rec), Rec{
			X: 1,
			S: "one",
			Next: &Rec{
				X: 2,
				S: "two",
				Next: &Rec{
					X: 3,
					S: "three",
				},
			},
		}},
	} {
		tb.Run(tc.N, func(tb *testing.T) {
			var d Decoder
			uns = map[unsafe.Pointer]unmarshaler{}

			_, err := d.Unmarshal([]byte(tc.D), 0, tc.X)
			if !assert.NoError(tb, err) {
				return
			}

			rres := reflect.ValueOf(tc.X).Elem()
			rexp := reflect.ValueOf(tc.E)

			for rres.Kind() == reflect.Pointer && !rres.IsNil() {
				rres = rres.Elem()
				rexp = rexp.Elem()
			}

			res := rres.Interface()
			exp := rexp.Interface()

			//	log.Printf("unmarshal\n`%s`\n%+v\n", tc.D, deref)
			tb.Logf("unmarshal\n`%s`\n%+v (%[2]T)", tc.D, res)

			if tc.E == nil {
				return
			}

			assert.Equal(tb, exp, res)
		})
	}
}

func ptr[T any](x T) *T {
	return &x
}
