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
	i := v.i
	b := v.buf
	if !v.parsed {
		_, err := skipValue(b, 0)
		if err != nil {
			return 0, err
		}
		v.parsed = true
	}
	i = skipSpaces(b, i)
	if i == len(b) {
		return Null, NewError(b, i, ErrExpectedValue)
	}
	switch b[i] {
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
