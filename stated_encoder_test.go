package json

import (
	"bytes"
	"testing"
)

func TestStatedEncoder(tb *testing.T) {
	e := NewStatedEncoder(nil)

	e.ObjStart()

	e.Key("a").String("value")

	e.Key("b").ArrStart().Int(1).Int(2).Int(3).ArrEnd()

	e.Key("c").ArrStart()
	e.Int(1)
	e.String("val")
	e.ArrStart()
	e.ObjStart().
		KeyString("c_a", "c_a_val").
		KeyInt("c_b", 10).
		ObjEnd()
	e.ObjStart().
		KeyString("c_c", "c_c_val").
		KeyInt("c_d", 20).
		ObjEnd()
	e.ArrEnd()
	e.Int64(64)
	e.ArrEnd()

	e.ObjEnd()

	b := e.Newline().Result()

	if !bytes.Equal([]byte(`{"a":"value","b":[1,2,3],"c":[1,"val",[{"c_a":"c_a_val","c_b":10},{"c_c":"c_c_val","c_d":20}],64]}`+"\n"), b) {
		tb.Errorf("wrong: %s", b)
	}
}

func TestStatedEncoderMulti(tb *testing.T) {
	e := NewStatedEncoder(nil)

	e.ObjStart().KeyInt("a", 1).ObjEnd()
	e.ObjStart().KeyInt("a", 2).ObjEnd()
	e.ObjStart().KeyInt("a", 3).ObjEnd()

	b := e.Newline().Result()

	if !bytes.Equal([]byte("{\"a\":1}\n{\"a\":2}\n{\"a\":3}\n"), b) {
		tb.Errorf("wrong: %s", b)
	}
}
