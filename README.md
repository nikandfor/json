[![Documentation](https://pkg.go.dev/badge/github.com/nikandfor/json)](https://pkg.go.dev/github.com/nikandfor/json?tab=doc)
[![Go workflow](https://github.com/nikandfor/json/actions/workflows/go.yml/badge.svg)](https://github.com/nikandfor/json/actions/workflows/go.yml)
[![CircleCI](https://circleci.com/gh/nikandfor/json.svg?style=svg)](https://circleci.com/gh/nikandfor/json)
[![codecov](https://codecov.io/gh/nikandfor/json/branch/master/graph/badge.svg)](https://codecov.io/gh/nikandfor/json)
[![GolangCI](https://golangci.com/badges/github.com/nikandfor/json.svg)](https://golangci.com/r/github.com/nikandfor/json)
[![Go Report Card](https://goreportcard.com/badge/github.com/nikandfor/json)](https://goreportcard.com/report/github.com/nikandfor/json)
![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/nikandfor/json?sort=semver)

# json

This is one more json library.
It's created to process unstructured json in a convenient and efficient way.

There is also some set of [jq](https://jqlang.github.io/jq/manual/) filters implemented on top of `json.Parser`.

## usage

`Parser` is stateless.
Most of the methods take source buffer and index where to start parsing and return a result and index where they stopped parsing.

```go
// Parsing single object.

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
```

```go
// Parsing jsonl: newline (or space, or comma) delimited values.

var err error // to not to shadow i in a loop
var p json.Parser
data := []byte(`"a", 2 3
["array"]
`)

for i := 0; i < len(data); i = p.SkipSpaces(data, i) { // eat trailing spaces and not try to read the value from string "\n"
	i, err = processOneObject(data, i) // do not use := here as it shadow i and loop will restart from the same index
	if err != nil {
		// ...
	}
}
```

## jq usage

jq package is a set of Filters that take data from one buffer, process it, and add result to another buffer.
Filters take destination buffer, source buffer, and starting index in a source buffer
and return updated destination buffer, ending index, and possible error.

Destination buffer is returned even in case of error.
This is mostly for avoiding allocs in case the buffer was grown but error happened.

```go
// Extract some inside value.

data := []byte(`{"key0":"skip it", "key1": {"next_key": ["array", null, {"obj":"val"}, "trailing element"]}}  "next"`)

f := jq.Selector{"key1", "next_key", 2} // string keys and int array indexes are supported

var res []byte // reusable buffer
var i int      // start index

// Most filters only parse single value and return index where the value ended.
// Use jq.ApplyToAll(f, res[:0], data, 0) to process all values in a buffer.
res, i, err := f.Apply(res[:0], data, i)
if err != nil {
	// i is an index in a source buffer where the error occured.
}

fmt.Printf("value: %s", res)                           // res ends on newline
fmt.Printf("final position: %d of %d\n", i, len(data)) // object was parsed to the end to be able to read next
_ = data                                               // but not the next value
```
