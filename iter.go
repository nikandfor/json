package json

// HasNext checks if you can read next element at the current object (or array).
// It also reads opening and closing brakets internally
//
// It moves position inside object (or array), so you'll get different result if you
// call it twice in a row. For example you have the following json
//		{"a": [{"b": "c"}, {"d": "e"}]}
//		    ^ - cursor is here (you've just read "a" and want to iterate over it's value)
//		// first time you call HasNext it returns true and moves cursor
//		{"a": [{"b": "c"}, {"d": "e"}]}
//		       ^ - here
//		// second time you call HasNext it returns true and moves cursor
//		{"a": [{"b": "c"}, {"d": "e"}]}
//		        ^ - here (inside b->c object)
//		// all subsequent calls will not move cursor anywhere until you read "b" key
// See example how to use it properly
func (r *Reader) HasNext() bool {
	//	log.Printf("HasNxt: %d+%d/%d '%c'", r.ref, r.i, r.end, r.b[r.i])
	//	defer func() {
	//		log.Printf("HasNxt: %d+%d/%d -> %v", r.ref, r.i, r.end, r_)
	//	}()

	var prev byte
start:
	for r.i < r.end {
		c := r.b[r.i]
		switch c {
		case ' ', '\t', '\n':
			r.i++
			continue
		case ',':
			r.i++
			prev = c
			continue
		case ':':
			r.i++
			continue
		case '{', '[':
			if prev == ',' || prev == '[' {
				return true
			}
			if prev != '{' {
				r.i++
				prev = c
				continue
			}
		case '}', ']':
			if prev == 0 || prev+2 == c {
				r.i++
				return false
			}
		case '"':
			return true
		case 't', 'f', 'n', '.', '-', '+':
			if prev != '{' {
				return true
			}
		default:
			if c >= '0' && c <= '9' {
				if prev != '{' {
					return true
				}
			}
		}
		r.err = errInvalidChar
		return false
	}
	if r.more() {
		goto start
	}
	return false
}
