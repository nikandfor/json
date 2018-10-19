package json

import (
	"io"
	"strings"
	"testing"
	"unicode/utf8"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCompareKey(t *testing.T) {
	v := WrapString(`"key_a":"res"`)
	ok := v.CompareKey([]byte("key_a"))
	assert.True(t, ok)
	assert.Equal(t, 7, v.i)
	assert.NoError(t, v.err)

	v = WrapString(`"key_a":"res"`)
	ok = v.CompareKey([]byte("key_b"))
	assert.False(t, ok)
	assert.Equal(t, 7, v.i)
	assert.NoError(t, v.err)

	v = WrapString(`"key_a":"res"`)
	ok = v.CompareKey([]byte("key"))
	assert.False(t, ok)
	assert.Equal(t, 7, v.i)
	assert.NoError(t, v.err)

	v = WrapString(`"key_a":"res"`)
	ok = v.CompareKey([]byte("key_long"))
	assert.False(t, ok)
	assert.Equal(t, 7, v.i)
	assert.NoError(t, v.err)
}

func TestSkipStrings(t *testing.T) {
	data := `"str_1""str_2" "str_3" "str_4""str_5"`
	v := WrapString(data)

	j := 0
	for v.Type() != None {
		t.Logf("iter %d: %2v/%2v '%s' %v", j, v.i, v.end, v.b, v.err)
		v.Skip()
		j++
	}

	err := v.Err()
	assert.True(t, err == nil || errors.Cause(err) == io.EOF)
	assert.Equal(t, 5, j)

	t.Logf("iter _: %2v/%2v '%s' %v", v.i, v.end, v.b, v.err)
}

func TestSkipArrays(t *testing.T) {
	data := `[] [] [][][]`
	v := WrapString(data)

	j := 0
	for v.Type() != None {
		t.Logf("iter %d: %2v/%2v '%s' %v", j, v.i, v.end, v.b, v.err)
		v.Skip()
		j++
	}

	assert.NoError(t, v.Err())
	assert.Equal(t, 5, j)

	t.Logf("iter _: %2v/%2v '%s' %v", v.i, v.end, v.b, v.err)
}

func TestSkipObjects(t *testing.T) {
	data := `{}{} {} {}{}`
	v := WrapString(data)

	j := 0
	for v.Type() != None {
		t.Logf("iter %d: %2v/%2v '%s' %v", j, v.i, v.end, v.b, v.err)
		v.Skip()
		j++
	}

	assert.NoError(t, v.Err())
	assert.Equal(t, 5, j)

	t.Logf("iter _: %2v/%2v '%s' %v", v.i, v.end, v.b, v.err)
}

func TestSkipObjectsNested(t *testing.T) {
	data := `{"a":{"b":{"c":{},"d":{}}},"e":{}}{}`
	v := WrapString(data)

	j := 0
	for v.Type() != None {
		t.Logf("iter %d: %2v/%2v '%s' %v", j, v.i, v.end, v.b, v.err)
		v.Skip()
		j++
	}

	assert.NoError(t, v.Err())
	assert.Equal(t, 2, j)

	t.Logf("iter _: %2v/%2v '%s' %v", v.i, v.end, v.b, v.err)
}

func TestSkipTopic(t *testing.T) {
	data := `{"id":155,"title":"Query working on \"Questions\" but not in \"Pulses\"","fancy_title":"Query working on &ldquo;Questions&rdquo; but not in &ldquo;Pulses&rdquo;","slug":"query-working-on-questions-but-not-in-pulses","posts_count":3,"reply_count":0,"highest_post_number":3,"image_url":null,"created_at":"2016-01-01T14:06:10.083Z","last_posted_at":"2016-01-08T22:37:51.772Z","bumped":true,"bumped_at":"2016-01-08T22:37:51.772Z","unseen":false,"pinned":false,"unpinned":null,"visible":true,"closed":false,"archived":false,"bookmarked":null,"liked":null,"views":72,"like_count":0,"has_summary":false,"archetype":"regular","last_poster_username":"agilliland","category_id":1,"pinned_globally":false,"posters":[{"extras":null,"description":"Original Poster","user_id":84},{"extras":null,"description":"Frequent Poster","user_id":73},{"extras":"latest","description":"Most Recent Poster","user_id":14}]},{"id":161,"title":"Pulses posted to Slack don't show question output","fancy_title":"Pulses posted to Slack don&rsquo;t show question output","slug":"pulses-posted-to-slack-dont-show-question-output","posts_count":2,"reply_count":0,"highest_post_number":2,"image_url":"/uploads/default/original/1X/9d2806517bf3598b10be135b2c58923b47ba23e7.png","created_at":"2016-01-08T22:09:58.205Z","last_posted_at":"2016-01-08T22:28:44.685Z","bumped":true,"bumped_at":"2016-01-08T22:28:44.685Z","unseen":false,"pinned":false,"unpinned":null,"visible":true,"closed":false,"archived":false,"bookmarked":null,"liked":null,"views":34,"like_count":0,"has_summary":false,"archetype":"regular","last_poster_username":"sameer","category_id":1,"pinned_globally":false,"posters":[{"extras":null,"description":"Original Poster","user_id":87},{"extras":"latest","description":"Most Recent Poster","user_id":1}]}`
	v := WrapString(data)

	j := 0
	for v.Type() != None {
		t.Logf("iter %d: %2v/%2v %v", j, v.i, v.end, v.err)
		v.Skip()
		j++
	}

	assert.NoError(t, v.Err())
	assert.Equal(t, 2, j)

	t.Logf("iter _: %2v/%2v '%s' %v", v.i, v.end, v.b, v.err)
}

func TestSkipNotAValue(t *testing.T) {
	data := `{"key": here}`
	v := WrapString(data)

	for v.Type() != None {
		v.Skip()
	}

	if err := v.Err(); assert.Error(t, err) {
		e := err.(Error)
		assert.Equal(t, 8, e.Pos())
	}
}

func TestReader(t *testing.T) {
	data := `{"a":{"b":[1,2,3],"c":"c_val"},"d":1.2,"e":3e-5} {}`
	t.Logf("data %d: '%s'", len(data), data)
	t.Logf("____   : '%s'", string("_123456789_123456789_123456789_123456789_123456789_123456789_")[:len(data)])

	r := strings.NewReader(data)
	v := NewReaderBufferSize(r, 1)

	j := 0
	for v.Type() != None {
		t.Logf("iter %d: %2v + %2v/%2v '%s' %v", j, v.ref, v.i, v.end, v.b, v.err)
		v.Skip()
		j++
	}

	assert.Error(t, v.Err(), io.EOF.Error())
	assert.Equal(t, 2, j)

	t.Logf("iter _: %2v + %2v/%2v '%s' %v", v.ref, v.i, v.end, v.b, v.err)
}

func TestGetObjects(t *testing.T) {
	data := `{"a":{"b":{"c":"d"},"e":{"f":"g"}}}`
	t.Logf("data %d: '%s'", len(data), data)
	t.Logf("____   : '%s'", string("_123456789_123456789_123456789_123456789_123456789_123456789_")[:len(data)])

	rd := strings.NewReader(data)
	r := NewReaderBufferSize(rd, 10)

	r.Get("a", "e", "f")
	assert.NoError(t, r.Err())
	assert.Equal(t, String, r.Type())

	t.Logf("iter _: %2v + %2v/%2v '%s' %v", r.ref, r.i, r.end, r.b, r.err)

	assert.Equal(t, "g", string(r.NextString()))

	t.Logf("iter _: %2v + %2v/%2v '%s' %v", r.ref, r.i, r.end, r.b, r.err)
}

func TestGetArrays(t *testing.T) {
	data := `[[[1,2,3],[4,5]],[6,7,[8,[9,10,11],12]]]`
	t.Logf("data %d: '%s'", len(data), data)
	t.Logf("____   : '%s'", string("_123456789_123456789_123456789_123456789_123456789_123456789_")[:len(data)])

	r := strings.NewReader(data)
	v := NewReaderBufferSize(r, 10)

	v.Get(1, 2, 1, 1)
	assert.NoError(t, v.Err())
	assert.Equal(t, Number, v.Type())

	t.Logf("iter _: %2v + %2v/%2v '%s' %v", v.ref, v.i, v.end, v.b, v.err)

	assert.Equal(t, "10", string(v.NextNumber()))

	t.Logf("iter _: %2v + %2v/%2v '%s' %v", v.ref, v.i, v.end, v.b, v.err)
}

func TestGet2(t *testing.T) {
	data := `{"a":{"b":[true,false,null],"c":false},"d":{"eee":[{"a":1,"c":{"val":"not_result"}},{"a":1,"b":[1,2,3],"c":{"val":"result"}}]},"e":null}`
	t.Logf("data %d: '%s'", len(data), data)
	t.Logf("____   : '%s'", string("_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_")[:len(data)])

	r := strings.NewReader(data)
	v := NewReaderBufferSize(r, 10)

	v.Get("d", "eee", 1, "c", "val")
	assert.NoError(t, v.Err())
	assert.Equal(t, String, v.Type())
	assert.Equal(t, "result", string(v.NextString()))

	t.Logf("iter _: %2v + %2v/%2v '%s' %v", v.ref, v.i, v.end, v.b, v.err)
}

func TestNextNumber(t *testing.T) {
	data := ` 10,11, 12 ,13 , 14 ,15`
	r := WrapString(data)

	for i := 0; i < 6; i++ {
		v, err := r.Int()
		assert.NoError(t, err)
		assert.Equal(t, 10+i, v)
	}

	assert.Equal(t, None, r.Type())
}

func TestSkipNumbers(t *testing.T) {
	data := ` 10,11, 12 ,13 , 14 ,15`
	r := WrapString(data)

	for i := 0; i < 6; i++ {
		assert.Equal(t, Number, r.Type())
		r.Skip()
		assert.NoError(t, r.Err())
	}

	assert.Equal(t, None, r.Type())
}

func TestBytesNumbers(t *testing.T) {
	data := ` 10,11, 12 ,13 , 14 ,15`
	r := WrapString(data)

	for i := 0; i < 6; i++ {
		assert.Equal(t, Number, r.Type())
		v := r.NextBytes()
		assert.Equal(t, []byte{'1', '0' + byte(i)}, v)
	}

	assert.NoError(t, r.Err())
	assert.Equal(t, None, r.Type())
}

func TestSkipStringUTF8(t *testing.T) {
	data := `"строка молока" {"ключ":"значение","массив":["один","日本語","три"]}`
	t.Logf("data %3d: '%s'", len(data), data)
	pad := nums(data)
	t.Logf("____ %3d: '%s'", len(pad), pad)

	r := strings.NewReader(data)
	v := NewReaderBufferSize(r, 5)

	tp := v.Type()
	assert.Equal(t, String, tp)
	t.Logf("iter _: %2v + %2v/%2v '%s' %v", v.ref, v.i, v.end, v.b, v.err)

	j := 0
	for v.Type() != None {
		t.Logf("iter %d: %2v + %2v/%2v '%s' %v", j, v.ref, v.i, v.end, v.b, v.err)
		v.Skip()
		j++
	}

	assert.Error(t, v.Err(), io.EOF.Error())
	assert.Equal(t, 2, j)

	t.Logf("iter _: %2v + %2v/%2v '%s' %v", v.ref, v.i, v.end, v.b, v.err)
}

func nums(s string) []byte {
	pad := make([]byte, len(s))
	j := 0
	for i, r := range s {
		d := i % 10
		if d == 0 {
			pad[j] = '_'
		} else {
			pad[j] = '0' + (byte)(d)
		}
		j++
		if utf8.RuneLen(r) > 2 {
			pad[j] = ' '
			j++
		}
	}
	pad = pad[:j]
	return pad
}

func TestEscaping(t *testing.T) {
	data := `" \t\n" "лол" "\t кек"`
	r := WrapString(data)

	exp := []string{
		" \t\n", "лол", "\t кек",
	}
	for i := 0; r.HasNext(); i++ {
		t.Logf("iter %d: %2v + %2v/%2v '%s' %v  %v", i, r.ref, r.i, r.end, r.b, r.err, r.Type())
		s := r.NextString()
		//	t.Logf("str got: '%s'", s)
		assert.Equal(t, []byte(exp[i]), s, "for %d '%s'", i, exp[i])
	}
	t.Logf("iter _: %2v + %2v/%2v '%s' %v", r.ref, r.i, r.end, r.b, r.err)

	assert.NoError(t, r.Err())
}

func TestSizeof(t *testing.T) {
	s := unsafe.Sizeof(Reader{})

	t.Logf("sizeof Reader{}: %d", s)
}
