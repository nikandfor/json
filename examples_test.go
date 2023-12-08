package json_test

import (
	"fmt"

	"nikand.dev/go/json"
)

func ExampleDecoder() {
	var d json.Decoder
	data := []byte(`{"key": "value", "another": 1234}`)

	i := 0 // initial position
	i, err := d.Enter(data, i, json.Object)
	if err != nil {
		// not an object
	}

	var key []byte // to not to shadow i and err in a loop

	// extracted values
	var value, another []byte

	for d.ForMore(data, &i, json.Object, &err) {
		key, i, err = d.Key(data, i) // key decodes a string but don't decode '\n', '\"', '\xXX' and others
		if err != nil {
			// ...
		}

		switch string(key) {
		case "key":
			value, i, err = d.DecodeString(data, i, value[:0]) // reuse value buffer if we are in a loop or something
		case "another":
			another, i, err = d.Raw(data, i)
		default: // skip additional keys
			i, err = d.Skip(data, i)
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

func ExampleDecoder_multipleValues() {
	var err error // to not to shadow i in a loop
	var d json.Decoder
	data := []byte(`"a", 2 3
["array"]
`)

	processOneObject := func(data []byte, st int) (int, error) {
		raw, i, err := d.Raw(data, st)

		fmt.Printf("value: %s\n", raw)

		return i, err
	}

	for i := d.SkipSpaces(data, 0); i < len(data); i = d.SkipSpaces(data, i) { // eat trailing spaces and not try to read the value from string "\n"
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
