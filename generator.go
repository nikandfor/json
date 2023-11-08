package json

import "unicode/utf8"

type (
	Generator struct{}
)

func (g *Generator) EncodeString(w, s []byte) []byte {
	w = append(w, '"')
	w = g.EncodeStringContent(w, s)
	w = append(w, '"')

	return w
}

func (g *Generator) EncodeStringContent(w, s []byte) []byte {
	done := 0

	for i := 0; i < len(s); {
		switch s[i] {
		case '"', '\\', '\n', '\t', '\v':
			w = append(w, s[done:i]...)

			switch s[i] {
			case '"', '\\':
				w = append(w, '\\', s[i])
			case '\n':
				w = append(w, '\\', 'n')
			case '\t':
				w = append(w, '\\', 't')
			case '\v':
				w = append(w, '\\', 'v')
			}

			i++
			done = i
		default:
			_, wid := utf8.DecodeRune(s[i:])
			i += wid
		}
	}

	w = append(w, s[done:]...)

	return w
}
