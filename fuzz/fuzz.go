package json

import "github.com/nikandfor/json"

func Fuzz(d []byte) int {
	_, err := json.Parse(d)
	if err != nil {
		return 0
	}
	return 1
}
