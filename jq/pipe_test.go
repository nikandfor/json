package jq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipeQuery(t *testing.T) {
	data := `{"a":"b","c":4,"d":["e",null,true,false,{"f":3.4}]}`

	f := NewPipe(
		NewQuery("d"),
		NewQuery(4),
		NewQuery("f"),
	)

	b, i, state, err := f.Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Nil(t, state)
	assert.Len(t, data, i)
	assert.Equal(t, "3.4", string(b))

	//	assert.True(t, cap(f.Bufs[0]) != 0)
	//	assert.True(t, cap(f.Bufs[1]) != 0)
}

func TestPipeComma(t *testing.T) {
	data := `null`

	var err error
	var state State
	var i int
	var b []byte

	f := NewPipe(
		NewComma(Literal(`"a"`), Literal(`"b"`)),
		NewComma(
			&Array{Filter: NewComma(Dot{}, Literal(`1`))},
			&Array{Filter: NewComma(Dot{}, Literal(`2`))},
		),
	)

	for j, exp := range []string{
		`["a",1]`,
		`["a",2]`,
		`["b",1]`,
		`["b",2]`,
	} {
		b, i, state, err = f.Next(b[:0], []byte(data), i, state)
		if !assert.NoError(t, err) ||
			!assert.True(t, (state == nil) == (j == 3), "index %d  state %v", j, state) ||
			//	assert.Equal(t, 0, i)
			!assert.Equal(t, exp, string(b)) {
			t.Logf("index: %v", j)
		}
	}

	assert.Nil(t, state)
}

func BenchmarkPipeQuery(b *testing.B) {
	b.ReportAllocs()

	data := []byte(`{"a":"b","c":4,"d":["e",null,true,false,{"f":3.4}]}`)

	var err error
	//	var state State
	//	var st int
	var buf []byte

	f := NewPipe(
		NewQuery("d"),
		NewQuery(4),
		NewQuery("f"),
	)

	for i := 0; i < b.N; i++ {
		buf, _, _, err = f.Next(buf[:0], data, 0, nil)
		if err != nil {
			b.Errorf("pipe: %v", err)
		}
	}
}

func BenchmarkPipe(b *testing.B) {
	b.ReportAllocs()

	data := []byte(`null`)

	var err error
	var state State
	var st int
	var buf []byte

	f := NewPipe(
		NewComma(Literal(`"a"`), Literal(`"b"`)),
		NewComma(
			&Array{Filter: NewComma(Dot{}, Literal(`1`))},
			&Array{Filter: NewComma(Dot{}, Literal(`2`))},
		),
	)

	for i := 0; i < b.N; i++ {
		buf, st, state, err = f.Next(buf, data, st, state)
		if err != nil {
			b.Errorf("pipe: %v", err)
		}

		if state == nil {
			st = 0
			buf = buf[:0]
		}
	}
}
