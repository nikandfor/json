package nikandjson

import "github.com/nikandfor/json"

func Fuzz(d []byte) int {
	r := json.Wrap(d)
	r.Skip()
	if r.Err() != nil {
		return 0
	}
	return 1
}
