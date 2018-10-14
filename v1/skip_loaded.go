package json

////go:noescape

func skipString(b []byte, s int) (int, error) {
	if b[s] != '"' {
		return s, NewError(b, s, ErrUnexpectedChar)
	}
	esc := false
	for i, c := range b[s+1:] {
		if c == '\\' {
			if esc {
				esc = false
				continue
			}
			if s+1+i+1 >= len(b) {
				return i + 1, NewError(b, i+1, ErrUnexpectedChar)
			}

			cn := b[s+1+i+1]
			if cn == '"' || cn == '\\' {
				esc = true
			}

			switch cn {
			case '"', '\\', '/', '\'', 'b', 'f', 'n', 'r', 't':
				continue
			default:
				return i + 1, NewError(b, i+1, ErrUnexpectedChar)
			}
		}
		if c == '"' {
			if esc {
				esc = false
				continue
			}
			return s + i + 2, nil
		}
		//	if c < 0x20 || c >= 0x80 {
		//	//	if !unicode.IsPrint(rune(c)) {
		//		return s + 1 + i, NewError(b, s+1+i, ErrUnexpectedChar)
		//	}
	}
	return len(b), NewError(b, s+len(b), ErrUnexpectedEnd)
}

func decodeString(b []byte, s int) ([]byte, error) {
	if s == len(b) {
		return nil, NewError(b, s, ErrUnexpectedEnd)
	}
	if b[s] != '"' {
		return nil, NewError(b, s, ErrUnexpectedChar)
	}
	esc := false
	var r int
	ref := s + 1
	for i, c := range b[ref:] {
		if c == '\\' {
			if esc {
				esc = false
				continue
			}
			if ref+i+1 >= len(b) {
				return nil, NewError(b, ref+i+1, ErrUnexpectedChar)
			}

			cn := b[ref+i+1]
			if cn == '"' || cn == '\\' {
				esc = true
			}

			switch cn {
			case '"', '\\', '/', '\'', 'b', 'f', 'n', 'r', 't':
				r = i
				break
			default:
				return nil, NewError(b, ref+i+1, ErrUnexpectedChar)
			}
		}
		if c == '"' {
			if esc {
				esc = false
				continue
			}
			return b[ref : s+i+2], nil
		}
		if c < 0x20 || c >= 0x80 {
			//	if !unicode.IsPrint(rune(c)) {
			return nil, NewError(b, ref+i, ErrUnexpectedChar)
		}
	}
	if r == 0 {
		return nil, NewError(b, s+len(b), ErrUnexpectedEnd)
	}

	ref = s + 1 + r
	res := make([]byte, r)
	copy(res, b[s+1:ref])
	for i, c := range b[ref:] {
		if esc {
			switch c {
			case '"', '\\', '/', '\'':
				res = append(res, c)
				continue
			case 'n':
				res = append(res, '\n')
			case 't':
				res = append(res, '\t')
			case 'f':
				res = append(res, '\f')
			case 'b':
				res = append(res, '\b')
			case 'r':
				res = append(res, '\r')
			default:
				return nil, NewError(b, ref+i, ErrUnexpectedChar)
			}
			esc = false
			continue
		}
		if c == '\\' {
			esc = true
			continue
		}
		if c == '"' {
			return res, nil
		}
		if c < 0x20 || c >= 0x80 {
			//	if !unicode.IsPrint(rune(c)) {
			return nil, NewError(b, ref+i, ErrUnexpectedChar)
		}
		res = append(res, c)
	}

	return res, nil
}
