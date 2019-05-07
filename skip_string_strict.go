// +build !unsafestrings

package json

import "unicode/utf8"

func (r *Reader) skipString(esc bool) {
	//	log.Printf("Skip stri %d+%d/%d", r.ref, r.i, r.end)
start:
	i := r.i
loop:
	for i < r.end {
		c := r.b[i]
		//	log.Printf("skip str0 %d+%d/%d '%c'", r.ref, r.i, r.end, c)
		switch {
		case c == '\\':
			i++
			if esc {
				esc = false
				continue
			}
			esc = true
		case c == '"':
			i++
			if esc {
				esc = false
				continue
			}
			r.i = i
			return
		case c < 0x80: // utf8.RuneStart
			//	log.Printf("skip stri %d+%d/%d '%c' (%d)", r.ref, i, r.end, c, c)
			i++
			continue
		default:
			if i+utf8.UTFMax > r.end {
				if !utf8.FullRune(r.b[i:r.end]) {
					break loop
				}
			}
			n, s := utf8.DecodeRune(r.b[i:])
			if n == utf8.RuneError {
				r.err = ErrEncoding
				return
			}
			//	log.Printf("skip rune %d+%d/%d '%c'", r.ref, i, r.end, n)
			i += s
		}
	}
	r.i = i
	if r.more() {
		goto start
	}
}
