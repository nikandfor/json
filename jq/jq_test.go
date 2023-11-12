package jq

import (
	"testing"

	"nikand.dev/go/json"
	"github.com/stretchr/testify/assert"
)

type (
	TestFilter [][]byte
)

func TestSimple(t *testing.T) {
	data := `{"a":"b","c":4,"d":["e",null,true,false,{"f":3.4}]}`

	b, i, _, err := Dot{}.Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, data, string(b))

	b, i, _, err = Index{"a"}.Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"b"`, string(b))

	b, i, _, err = Index{"c"}.Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `4`, string(b))

	b, i, _, err = Index{"non"}.Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `null`, string(b))

	b, i, _, err = Index{"d", 2}.Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `true`, string(b))

	b, i, _, err = Index{"d", 4, "f"}.Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `3.4`, string(b))
}

func TestComma(t *testing.T) {
	data := `{"a":"b","c":4,"d":["e",null,true,false,{"f":3.4}]}`

	f := NewComma(
		Index{"a"},
		Index{"c"},
		Index{"d", 0},
	)

	var state State

	b, i, state, err := f.Next(nil, []byte(data), 0, state)
	assert.NoError(t, err)
	assert.NotNil(t, state)
	//	assert.Equal(t, len(data), i)
	assert.Equal(t, `"b"`, string(b))

	b, i, state, err = f.Next(nil, []byte(data), i, state)
	assert.NoError(t, err)
	assert.NotNil(t, state)
	//	assert.Equal(t, len(data), i)
	assert.Equal(t, `4`, string(b))

	b, i, state, err = f.Next(nil, []byte(data), i, state)
	assert.NoError(t, err)
	assert.Nil(t, state)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"e"`, string(b))
}

func TestCommaPipe(t *testing.T) {
	data := `{"a":"b","c":4,"d":["e",null,true,false,{"f":3.4}]}`

	f := NewComma(
		NewPipe(Index{"a"}),
		NewPipe(Index{"c"}),
		NewPipe(Index{"d"}, Index{0}),
	)

	b, i, state, err := f.Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.NotNil(t, state)
	//	assert.Equal(t, len(data), i)
	assert.Equal(t, `"b"`, string(b))

	b, i, state, err = f.Next(nil, []byte(data), i, state)
	assert.NoError(t, err)
	assert.NotNil(t, state)
	//	assert.Equal(t, len(data), i)
	assert.Equal(t, `4`, string(b))

	b, i, state, err = f.Next(nil, []byte(data), i, state)
	assert.NoError(t, err)
	assert.Nil(t, state)
	assert.Equal(t, len(data), i)
	assert.Equal(t, `"e"`, string(b))
}

func TestPipeIndex(t *testing.T) {
	data := `{"a":"b","c":4,"d":["e",null,true,false,{"f":3.4}]}`

	f := NewPipe(
		Index{"d"},
		Index{4},
		Index{"f"},
	)

	b, i, state, err := f.Next(nil, []byte(data), 0, nil)
	assert.NoError(t, err)
	assert.Nil(t, state)
	assert.Equal(t, len(data), i)
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

func BenchmarkPipeIndex(b *testing.B) {
	b.ReportAllocs()

	data := []byte(`{"a":"b","c":4,"d":["e",null,true,false,{"f":3.4}]}`)

	var err error
	//	var state State
	//	var st int
	var buf []byte

	f := NewPipe(
		Index{"d"},
		Index{4},
		Index{"f"},
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

func BenchmarkNextAll(b *testing.B) {
	b.ReportAllocs()

	data := []byte(`[1,2,3,4,5,6]`)

	var err error
	var state State
	var st int
	var buf []byte

	f := Iter{}

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

func (f TestFilter) Next(w, r []byte, st int, state State) (_ []byte, i int, _ State, err error) {
	var p json.Parser

	i, err = p.Skip(r, st)
	if err != nil {
		return w, i, state, err
	}

	ss, _ := state.(int)
	if ss == len(f) {
		return w, i, nil, nil
	}

	w = append(w, f[ss]...)

	ss++
	state = ss

	return w, st, state, nil
}
