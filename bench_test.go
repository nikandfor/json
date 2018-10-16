package json

import "testing"

import "github.com/nikandfor/json/benchmarks"

func BenchmarkGetUnmarshal(b *testing.B) {
	b.ReportAllocs()

	data := []byte(`{
  "person": {
    "id": "d50887ca-a6ce-4e59-b89f-14f0b5d03b03",
    "name": {
      "fullName": "Leonid Bugaev",
      "givenName": "Leonid",
      "familyName": "Bugaev"
    },
    "email": "leonsbox@gmail.com",
    "gender": "male",
    "location": "Saint Petersburg, Saint Petersburg, RU",
  }
}`)

	type FullName struct {
		GivenName  string
		FamilyName string
	}
	var f FullName
	var r Reader

	for i := 0; i < b.N; i++ {
		r.Reset(data)
		// Get only small needed subobject and unmarshal it
		r.Get("person", "name").Unmarshal(&f)
	}
}

func skipString(b []byte, i int) int {
	i++
	for b[i] != '"' {
		i++
	}
	i++
	return i
}

func BenchmarkRawLoopMediumFast(b *testing.B) {
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
				r.skipString(true)
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

/*
func BenchmarkRawLoopMediumStrict(b *testing.B) {
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
				r.skipStringStrict()
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
*/

func BenchmarkSkipMedium(b *testing.B) {
	b.ReportAllocs()
	data := go_benchmark.MediumFixture
	var r Reader
	for i := 0; i < b.N; i++ {
		r.Reset(data)
		r.Skip()
	}
}

func BenchmarkSkip3(b *testing.B) {
	b.ReportAllocs()
	data := []byte("truetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetruetrue")
	r := &Reader{b: data, end: len(data)}
	for i := 0; i < b.N; i++ {
		for r.i < r.end {
			r.i++
			r.skip3('r', 'u', 'e')
		}
	}
}
