package json

import (
	stdjson "encoding/json"
	"testing"

	"github.com/nikandfor/json"
)

func FuzzStringEncodeDecode(f *testing.F) {
	f.Add([]byte(`abcdef`))
	f.Add([]byte(`\"/'`))
	f.Add([]byte("\t\b\r\n\f"))
	f.Add([]byte("\x13"))
	f.Add([]byte("\xf8"))

	var g json.Generator
	var p json.Parser

	//	g.ASCII = true

	f.Fuzz(func(t *testing.T, s []byte) {
		//	if bytes.Contains(s, []byte(`\x`)) {
		//		t.SkipNow()
		//	}

		b := g.EncodeString(nil, s)

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

	var p json.Parser

	f.Fuzz(func(t *testing.T, data []byte) {
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
			t.Errorf("decode: %v", err)
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
