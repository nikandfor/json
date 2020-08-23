// +build ignore

package nikandjson

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"unicode/utf8"

	"github.com/nikandfor/json"
)

func FuzzUnicode(d []byte) int {
	if !utf8.Valid(d) {
		return -1
	}

	data, err := json.Marshal(string(d))
	if err != nil {
		panic(fmt.Sprintf("marshal: %v", err))
	}

	var res string
	err = json.Unmarshal(data, &res)
	if err != nil {
		panic(fmt.Sprintf("unmarshal: %v", err))
	}

	if !bytes.Equal(d, []byte(res)) {
		panic(fmt.Sprintf("unequal %q != %q\n%v", d, res, hex.Dump(data)))
	}

	return 1
}
