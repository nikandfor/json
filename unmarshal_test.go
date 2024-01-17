package json

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshal(tb *testing.T) {
	var d Decoder

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
	} {
		tb.Run(tc.N, func(tb *testing.T) {
			_, err := d.Unmarshal([]byte(tc.D), 0, tc.X)
			if !assert.NoError(tb, err) {
				return
			}

			deref := reflect.ValueOf(tc.X).Elem().Interface()

			//	log.Printf("unmarshal\n`%s`\n%+v\n", tc.D, deref)
			tb.Logf("unmarshal\n`%s`\n%+v", tc.D, deref)

			if tc.E == nil {
				return
			}

			assert.Equal(tb, tc.E, deref)
		})
	}
}

func ptr[T any](x T) *T {
	return &x
}
