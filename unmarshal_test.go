package json

import (
	"encoding/json"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

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

		Arr   [3]int
		Slice []Arr

		SlArr struct {
			Slice Slice `json:"slice"`
			Arr   Arr   `json:"arr"`
		}

		Raw struct {
			Raw RawMessage `json:"raw"`
		}

		RawPtr struct {
			Raw *RawMessage `json:"raw"`
		}
	)

	var d Decoder

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

		{"ArrayShort", `[1,2]`, new([3]int), [3]int{1, 2, 0}},
		{"ArrayEq", `[1,2,3]`, new([3]int), [3]int{1, 2, 3}},
		{"ArrayLong", `[1,2,3,4]`, new([3]int), [3]int{1, 2, 3}},

		{"Slice", `[1,2,3]`, new([]int), []int{1, 2, 3}},

		{"SlArr", `{"slice":[[1,2,3],[4,5,6]], "arr": [7,8,9]}`, new(SlArr), SlArr{Slice: Slice{Arr{1, 2, 3}, Arr{4, 5, 6}}, Arr: Arr{7, 8, 9}}},

		{"[]byte", `"YWIgMTIgW10="`, new([]byte), []byte("ab 12 []")},

		{"RawMsg", `{"a":"b"}`, new(RawMessage), RawMessage(`{"a":"b"}`)},
		{"StdRawMsg", `{"a":"b"}`, new(json.RawMessage), json.RawMessage(`{"a":"b"}`)},

		{"Raw", `{"raw":{"a":"b"}}`, new(Raw), Raw{Raw: RawMessage(`{"a":"b"}`)}},
		{"RawPtr", `{"raw":{"a":"b"}}`, new(RawPtr), RawPtr{Raw: ptr(RawMessage(`{"a":"b"}`))}},
	} {
		tc := tc

		tb.Run(tc.N, func(tb *testing.T) {
			uns = map[unsafe.Pointer]unmarshaler{}

			i, err := d.Unmarshal([]byte(tc.D), 0, tc.X)
			if !assert.NoError(tb, err) {
				return
			}

			assert.Equal(tb, len(tc.D), i)

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
