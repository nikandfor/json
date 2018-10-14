// +build !amd64

package json

func skipSpaces(b []byte, s int) int {
	if s == len(b) {
		return s
	}
	for i, c := range b[s:] {
		switch c {
		case ' ', '\t', '\n':
			continue
		default:
			return s + i
		}
	}
	return len(b)
}
