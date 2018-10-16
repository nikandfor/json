package json

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
