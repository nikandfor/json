package json

type Type int

const (
	Null Type = iota
	Bool
	Number
	String
	Array
	Object
)

func (v *Value) Type() (Type, error) {
	if v.i == v.end {
		return Null, ErrExpectedValue
	}
	switch v.buf[v.i] {
	case '[':
		return Array, nil
	case '{':
		return Object, nil
	case '"':
		return String, nil
	case 't':
		return Bool, nil
	case 'f':
		return Bool, nil
	case 'n':
		return Null, nil
	default:
		return Number, nil
	}
}
