package fuzz

import (
	stdjson "encoding/json"
	"testing"

	"nikand.dev/go/json2"
)

func FuzzSkip(f *testing.F) {
	f.Add([]byte(" \t\n"))
	f.Add([]byte(`null`))
	f.Add([]byte(`true`))
	f.Add([]byte(`false`))
	f.Add([]byte(`NaN`))
	f.Add([]byte(`Inf`))
	f.Add([]byte(`-Inf`))
	f.Add([]byte(`Infinity`))
	f.Add([]byte(`-Infinity`))
	f.Add([]byte(`1`))
	f.Add([]byte(`1.2567`))
	f.Add([]byte(`-5e+123123`))
	f.Add([]byte(`6p-10`))
	f.Add([]byte(`"abc"`))
	f.Add([]byte(`"a\nbc"`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`[1,"str",null,true]`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"a":"b","c":4}`))

	var p json2.Iterator

	f.Fuzz(func(t *testing.T, b []byte) {
		if !stdjson.Valid(b) {
			return
		}

		i, err := p.Skip(b, 0)
		if err != nil {
			t.Errorf("skip: %v", err)
		}

		i = p.SkipSpaces(b, i)
		if i != len(b) {
			t.Errorf("uncomplete read: %v / %v", i, len(b))
		}

		if t.Failed() {
			t.Logf("input: %s", b)
		}
	})
}

func FuzzStringEncodeDecode(f *testing.F) {
	f.Add([]byte(`abcdef`))
	f.Add([]byte(`\"/'`))
	f.Add([]byte("\t\b\r\n\f"))
	f.Add([]byte("\x13"))
	f.Add([]byte("\xf8"))

	var g json2.Emitter
	var p json2.Iterator

	//	g.ASCII = true

	f.Fuzz(func(t *testing.T, s []byte) {
		//	if bytes.Contains(s, []byte(`\x`)) {
		//		t.SkipNow()
		//	}

		b := g.AppendString(nil, s)

		st := len(b)
		b, i, err := p.DecodeString(b, 0, b)
		if err != nil {
			t.Errorf("decode: %v", err)
		}
		if i != st {
			t.Errorf("parsed %d, expected %d", i, st)
		}

		//	if bytes.Equal(s, b[st:]) {
		//		return
		//	}

		data, err := stdjson.Marshal(string(s))
		if err != nil {
			t.Errorf("stdjson.Marshal: %v", err)
			return
		}

		var res string

		err = stdjson.Unmarshal(data, &res)
		if err != nil {
			t.Errorf("stdjson.Unmarshal: %v", err)
			return
		}

		if res != string(b[st:]) /*|| string(data) != string(b[:st])*/ {
			t.Errorf("not equal results")
		}

		if t.Failed() {
			t.Errorf("nik/json %q -> %s -> %q\n%[1]x -> %x -> %x", s, b[:st], b[st:])
			t.Errorf("stdjson  %q -> %s -> %q\n%[1]x -> %x -> %x", s, data, res)
		}

		{
			var res2 string

			err = stdjson.Unmarshal(b[:st], &res2)
			if err != nil {
				t.Errorf("stdjson.Unmarshal2: %v", err)
				return
			}

			b2, i, err := p.DecodeString(data, 0, nil)
			if err != nil {
				t.Errorf("decode2: %v", err)
			}
			if i != len(data) {
				t.Errorf("parsed2 %d, expected %d", i, st)
			}

			if res2 != string(b2) {
				t.Errorf("cross decode not equal\nnik -> std: %q\nstd -> nik: %q", res2, b2)
			}
		}
	})
}

func FuzzStringDecode(f *testing.F) {
	f.Add([]byte(`"abcdef"`))
	f.Add([]byte(`"\\\"/'"`))
	f.Add([]byte(`"\ufffd"`))
	f.Add([]byte(`"\u2028"`))

	var p json2.Iterator

	f.Fuzz(func(t *testing.T, data []byte) {
		tp, _, err := p.Type(data, 0)
		if err != nil || tp != json2.String {
			t.SkipNow()
		}

		b, i, err := p.DecodeString(data, 0, nil)

		end := len(data)
		if err == nil {
			end = i
		}

		var res string

		err1 := stdjson.Unmarshal(data[:end], &res)

		if err != nil && err1 != nil {
			t.SkipNow()
		}

		if err != nil {
			t.Errorf("decode: %v  (i %x)", err, i)
		}

		if err1 != nil {
			t.Logf("std decode: %v", err1)
			t.SkipNow()
		}

		if res != string(b) {
			t.Errorf("different results")
		}

		if t.Failed() {
			t.Logf("input: %q  [%[1]s]\nstd:   %q\nnik:   %q", data, res, b)
			t.Logf("input: %x\nstd:   %x\nnik:   %x", data, res, b)
		}
	})
}
