package json_test

import (
	"fmt"

	"github.com/nikandfor/json"
)

func ExampleParser() {
	var p json.Parser
	data := []byte(`{"key": "value", "another": 1234}`)

	i := 0 // initial position
	i, err := p.Enter(data, i, json.Object)
	if err != nil {
		// not an object
	}

	var key []byte // to not to shadow i and err in a loop

	// extracted values
	var value, another []byte

	for p.ForMore(data, &i, json.Object, &err) {
		key, i, err = p.Key(data, i) // key decodes a string but don't decode '\n', '\"', '\xXX' and others
		if err != nil {
			// ...
		}

		switch string(key) {
		case "key":
			value, i, err = p.DecodeString(data, i, value[:0]) // reuse value buffer if we are in a loop or something
		case "another":
			another, i, err = p.Raw(data, i)
		default: // skip additional keys
			i, err = p.Skip(data, i)
		}

		// check error for all switch cases
		if err != nil {
			// ...
		}
	}
	if err != nil {
		// ForMore error
	}

	fmt.Printf("key: %s\nanother: %s\n", value, another)

	// Output:
	// key: value
	// another: 1234
}

func ExampleParser_multipleValues() {
	var err error // to not to shadow i in a loop
	var p json.Parser
	data := []byte(`"a", 2 3
["array"]
`)

	processOneObject := func(data []byte, st int) (int, error) {
		raw, i, err := p.Raw(data, st)

		fmt.Printf("value: %s\n", raw)

		return i, err
	}

	for i := 0; i < len(data); i = p.SkipSpaces(data, i) { // eat trailing spaces and not try to read the value from string "\n"
		i, err = processOneObject(data, i) // do not use := here as it shadow i and loop will restart from the same index
		if err != nil {
			// ...
		}
	}

	// Output:
	// value: "a"
	// value: 2
	// value: 3
	// value: ["array"]
}
