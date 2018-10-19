package json

import (
	"io"
	"testing"

	"github.com/nikandfor/json/benchmarks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestIterArray(t *testing.T) {
	//	data := `{"a":{"b":[true,false,null],"c":false},"d":{"eee":[{"a":1,"c":{"val":"not_result"}},{"a":1,"b":[1,2,3],"c":{"val":"result"}}]},"e":null}`
	data := `[]`
	r := WrapString(data)
	assert.False(t, r.HasNext())
	assert.NoError(t, r.Err())
	assert.Equal(t, None, r.Type())

	data = `["a","d","e"]`
	t.Logf("data %d: '%s'", len(data), data)
	t.Logf("____   : '%s'", string("_123456789_123456789_123456789_")[:len(data)])

	r = WrapString(data)

	keys := []string{"a", "d", "e"}
	j := 0
	for r.HasNext() {
		ok := r.CompareKey([]byte(keys[j]))
		assert.True(t, ok, "expected key: %v", keys[j])
		j++
	}

	assert.NoError(t, r.Err())
	assert.Equal(t, None, r.Type())
}

func TestIterObject(t *testing.T) {
	//	data := `{"a":{"b":[true,false,null],"c":false},"d":{"eee":[{"a":1,"c":{"val":"not_result"}},{"a":1,"b":[1,2,3],"c":{"val":"result"}}]},"e":null}`
	data := `{}`
	r := WrapString(data)
	assert.False(t, r.HasNext())
	assert.NoError(t, r.Err())
	assert.Equal(t, None, r.Type())

	data = `{"a":1,"d":2,"e":3}`
	t.Logf("data %d: '%s'", len(data), data)
	t.Logf("____   : '%s'", string("_123456789_123456789_123456789_")[:len(data)])

	r = WrapString(data)

	keys := []string{"a", "d", "e"}
	j := 0
	for r.HasNext() {
		ok := r.CompareKey([]byte(keys[j]))
		assert.True(t, ok, "expected key: %v", keys[j])
		r.Skip()
		j++
	}

	err := r.Err()
	assert.True(t, err == nil || errors.Cause(err) == io.EOF)
	assert.Equal(t, None, r.Type())
}

func TestIterComplex(t *testing.T) {
	data := `{"a":{"b":[true,false,null],"c":false},"d":{"eee":[{"a":1,"c":{"val":"not_result"}},{"a":1,"b":[1,2,3],"c":{"val":"result"}}],"e2":null},"e":null}`
	t.Logf("data %3d: '%s'", len(data), data)
	t.Logf("____    : '%s'", string("_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_")[:len(data)])

	r := WrapString(data)

	keys := []string{"a", "d", "e"}
	sub := map[string][]string{"a": {"b", "c"}, "d": {"eee", "e2"}}
	j := 0
	for r.HasNext() {
		ok := r.CompareKey([]byte(keys[j]))
		assert.True(t, ok, "expected key: %v", keys[j])
		if sub, ok := sub[keys[j]]; ok {
			j := 0
			for r.HasNext() {
				ok := r.CompareKey([]byte(sub[j]))
				assert.True(t, ok, "expected subkey: %v", sub[j])
				r.Skip()
				j++
			}
		} else {
			r.Skip()
		}
		j++
	}

	assert.NoError(t, r.Err())
	assert.Equal(t, None, r.Type())
}

var smallLarge = []byte(`{"topics":{"topics":[{"id":8},{"id":169},{"id":168}]}}`)

func TestSmallLargeCount(t *testing.T) {
	data := smallLarge
	//	data = []byte(go_benchmark.LargeFixture)
	t.Logf("data %3d: '%s'", len(data), data)
	t.Logf("____    : '%s'", string("_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_123456789_")[:len(data)])
	w := Wrap(data)
	count := 0
	for w.HasNext() {
		ok := w.CompareKey([]byte("topics"))
		if !ok {
			w.Type()
			w.Skip()
			continue
		}
		for w.HasNext() {
			ok := w.CompareKey([]byte("topics"))
			if !ok {
				w.Skip()
				continue
			}
			for w.HasNext() {
				count++
				t.Logf("topic: %s", w.NextAsBytes())
			}
		}
	}
	assert.Equal(t, 3, count)
}

func TestLargeCount(t *testing.T) {
	data := []byte(go_benchmark.LargeFixture)
	w := Wrap(data)
	count := 0
	for w.HasNext() {
		ok := w.CompareKey([]byte("topics"))
		if !ok {
			w.Type()
			w.Skip()
			continue
		}
		for w.HasNext() {
			ok := w.CompareKey([]byte("topics"))
			if !ok {
				w.Skip()
				continue
			}
			for w.HasNext() {
				count++
				//	t.Logf("topic: %s", w.NextBytes())
				w.Skip()
			}
		}
	}
	assert.Equal(t, 30, count)
}

func TestIterTopic(t *testing.T) {
	data := `{"id":155,"title":"Query working on \"Questions\" but not in \"Pulses\"","fancy_title":"Query working on &ldquo;Questions&rdquo; but not in &ldquo;Pulses&rdquo;","slug":"query-working-on-questions-but-not-in-pulses","posts_count":3,"reply_count":0,"highest_post_number":3,"image_url":null,"created_at":"2016-01-01T14:06:10.083Z","last_posted_at":"2016-01-08T22:37:51.772Z","bumped":true,"bumped_at":"2016-01-08T22:37:51.772Z","unseen":false,"pinned":false,"unpinned":null,"visible":true,"closed":false,"archived":false,"bookmarked":null,"liked":null,"views":72,"like_count":0,"has_summary":false,"archetype":"regular","last_poster_username":"agilliland","category_id":1,"pinned_globally":false,"posters":[{"extras":null,"description":"Original Poster","user_id":84},{"extras":null,"description":"Frequent Poster","user_id":73},{"extras":"latest","description":"Most Recent Poster","user_id":14}]},{"id":161,"title":"Pulses posted to Slack don't show question output","fancy_title":"Pulses posted to Slack don&rsquo;t show question output","slug":"pulses-posted-to-slack-dont-show-question-output","posts_count":2,"reply_count":0,"highest_post_number":2,"image_url":"/uploads/default/original/1X/9d2806517bf3598b10be135b2c58923b47ba23e7.png","created_at":"2016-01-08T22:09:58.205Z","last_posted_at":"2016-01-08T22:28:44.685Z","bumped":true,"bumped_at":"2016-01-08T22:28:44.685Z","unseen":false,"pinned":false,"unpinned":null,"visible":true,"closed":false,"archived":false,"bookmarked":null,"liked":null,"views":34,"like_count":0,"has_summary":false,"archetype":"regular","last_poster_username":"sameer","category_id":1,"pinned_globally":false,"posters":[{"extras":null,"description":"Original Poster","user_id":87},{"extras":"latest","description":"Most Recent Poster","user_id":1}]}`
	v := WrapString(data)

	docs := 0
	for v.Type() != None {
		j := 0
		for v.HasNext() {
			t.Logf("%d. %s", j, v.NextAsBytes())
			v.Skip()
			j++
		}
		assert.Equal(t, 28, j)
		docs++
	}
	assert.Equal(t, 2, docs)

	t.Logf("iter _: %2v/%2v '%s' %v", v.i, v.end, v.b, v.err)
}
