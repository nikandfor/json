package json

////go:noescape

func skipSpaces(b []byte, s int) int {
	if s == len(b) {
		return s
	}
	for i, c := range b[s:] {
		switch c {
		case ' ', '\n', '\t', '\v', '\r':
			continue
		default:
			return s + i
		}
	}
	return s + len(b)
}

////go:noescape

func skipString(b []byte, s int) (int, error) {
	if b[s] != '"' {
		return s, NewError(b, s, ErrUnexpectedChar)
	}
	esc := false
	for i, c := range b[s+1:] {
		if c == '\\' && !esc {
			esc = true
			continue
		}
		if c == '"' && !esc {
			return s + i + 2, nil
		}
		esc = false
	}
	return s + len(b), NewError(b, s+len(b), ErrUnexpectedEnd)
}
