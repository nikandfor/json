package json

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorSimple(t *testing.T) {
	data := `{"some":"text",123}`
	err := errors.New("some error message")

	t.Logf("%v", NewError([]byte(data), 37, 15, err))
	t.Logf("%#v", NewError([]byte(data), 37, 15, err))
	t.Logf("%+v", NewError([]byte(data), 37, 15, err))

	t.Logf("%#10v", NewError([]byte(data), 37, 15, err))

	assert.Equal(t, "parse error at pos 37: some error message", fmt.Sprintf("%v", NewError([]byte(data), 37, 15, err)))
	assert.Equal(t, "parse error at pos 37: some error message `"+`{"some":"text",123}`+"`", fmt.Sprintf("%#v", NewError([]byte(data), 37, 15, err)))
	assert.Equal(t, "parse error at pos 37: some error message\n"+
		`{"some":"text",123}`+"\n"+
		"_______________^___\n", fmt.Sprintf("%+v", NewError([]byte(data), 37, 15, err)))
}

func TestErrorEscapeTabs(t *testing.T) {
	r, p := escapeString([]byte("\t\n \t\nq\t\n \n\t"), 5)
	assert.Equal(t, 9, p)
	assert.Equal(t, []byte(`\t\n \t\nq\t\n \n\t`), r)
}

func TestErrorEscapeRus(t *testing.T) {
	r, p := escapeString([]byte("йыпqжир"), 3)
	assert.Equal(t, 3, p)
	assert.Equal(t, []byte(`йыпqжир`), r)
}
