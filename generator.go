package json

import "unicode/utf8"

type (
	Generator struct{}
)

const hexdigits = "0123456789abcdef"

// EncodeString encodes string replacing some symbols with escape sequences.
// It also adds quotes.
func (g *Generator) EncodeString(w, s []byte) []byte {
	w = append(w, '"')
	w = g.EncodeStringContent(w, s)
	w = append(w, '"')

	return w
}

// EncodeStringContent does the same as EncodeString but does not add quotes.
// It can be used to generate the string from multiple parts.
// Yet if a symbol designated to be escaped is split with parts
// it encodes each part of the symbol separately.
func (g *Generator) EncodeStringContent(w, s []byte) []byte {
	done := 0

	for i := 0; i < len(s); {
		switch s[i] {
		case '"', '\\', '/', '\n', '\r', '\t', '\f', '\b':
			w = append(w, s[done:i]...)

			switch s[i] {
			case '"', '\\', '/':
				w = append(w, '\\', s[i])
			case '\n':
				w = append(w, '\\', 'n')
			case '\r':
				w = append(w, '\\', 'r')
			case '\t':
				w = append(w, '\\', 't')
			case '\f':
				w = append(w, '\\', 'f')
			case '\b':
				w = append(w, '\\', 'b')
			}

			i++
			done = i
		default:
			r, wid := utf8.DecodeRune(s[i:])

			if s[i] >= 0x20 && r != utf8.RuneError {
				i += wid
				break
			}

			w = append(w, s[done:i]...)
			w = append(w, '\\', 'u', '0', '0', hexdigits[s[i]>>4], hexdigits[s[i]&0xf])

			i++
			done = i
		}
	}

	w = append(w, s[done:]...)

	return w
}
