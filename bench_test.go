package json

import "testing"

import "github.com/nikandfor/json/benchmarks"

func skipString(b []byte, i int) int {
	i++
	for b[i] != '"' {
		i++
	}
	i++
	return i
}

func BenchmarkRawLoop(b *testing.B) {
	b.ReportAllocs()
	data := go_benchmark.MediumFixture
	r := &Reader{b: data, end: len(data)}
	var l, p, q, n, d int
	for i := 0; i < b.N; i++ {
		l, p, q, n, d = 0, 0, 0, 0, 0
		for j := 0; j < len(data); j++ {
			c := data[j]
			switch c {
			case ' ', '\t', '\n':
				continue
			}
			switch c {
			case '{', '[':
				d++
			case '}', ']':
				d--
			case ':':
				l++
			case ',':
				p++
			case '"':
				q++
				r.i = j
				r.skipString()
				j = r.i
			//	j = skipString(data, j)
			case '+', '-':
				n++
			default:
				if c >= '0' && c <= '9' {
					n++
				}
			}
		}
		_, _, _, _, _ = l, p, q, n, d
	}
	//b.Logf("l %d, p %d, q %d, n %d, d %d", l, p, q, n, d)
}

func BenchmarkSkip(b *testing.B) {
	data := []byte("truetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetrue")
	r := &Reader{b: data, end: len(data)}
	for i := 0; i < b.N; i++ {
		for r.i < r.end {
			r.skip([]byte{'t', 'r', 'u', 'e'})
		}
	}
}

func BenchmarkSkip1(b *testing.B) {
	data := []byte("truetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetrue")
	r := &Reader{b: data, end: len(data)}
	for i := 0; i < b.N; i++ {
		var j int
		for j < r.end {
			j = r.skip1(j, []byte("true"))
		}
	}
}

func (r *Reader) skip1(i int, k []byte) int {
	j := 0
start:
	for i < r.end {
		if j == len(k) {
			return i
		}
		c := r.b[i]
		if c != k[j] {
			r.err = ErrError
			return i
		}
		j++
		i++
	}
	r.i = i
	if r.more() {
		i = r.i
		goto start
	}
	return i
}
