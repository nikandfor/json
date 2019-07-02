[![Documentation](https://godoc.org/github.com/nikandfor/json?status.svg)](http://godoc.org/github.com/nikandfor/json)
[![Build Status](https://travis-ci.com/nikandfor/json.svg?branch=master)](https://travis-ci.com/nikandfor/json)
[![CircleCI](https://circleci.com/gh/nikandfor/json.svg?style=svg)](https://circleci.com/gh/nikandfor/json)
[![codecov](https://codecov.io/gh/nikandfor/json/branch/master/graph/badge.svg)](https://codecov.io/gh/nikandfor/json)
[![GolangCI](https://golangci.com/badges/github.com/nikandfor/json.svg)](https://golangci.com/r/github.com/nikandfor/json)
[![Go Report Card](https://goreportcard.com/badge/github.com/nikandfor/json)](https://goreportcard.com/report/github.com/nikandfor/json)
![Project status](https://img.shields.io/badge/status-Developing-yellow.svg)

# json - Fast but Flexible JSON library for golang

This is one more json decoding/encoding library. Why should you use it? Because it's
* Fast
* Memory efficient
* Supports Unmarshal interface as well as raw data iterators
* Supports decoding/encoding from/to memory buffer as well as io.Reader/Writer
* Tested by unit tests, benchmarks and fuzzing
* It doesn't do what you didn't ask saving time and resources
* It can unmarshal struct and don't zero fields you didn't ask (which weren't present in data)

If you don't have [go2doc](https://chrome.google.com/webstore/detail/go2doc/mnpdpppgidppdhingkmlcmmgdjknecif) than use a link https://godoc.org/github.com/nikandfor/json

## Status
It's mostly done but I'm still may change something, including api.
Although it tested and even fuzzed (https://github.com/dvyukov/go-fuzz) it's used only at my projects by now, so some situations that are not crashes but getting unexpected results may happen. So error reports are welcome. If you miss some feature feel free to ask for it either.

## Examples

Conventional Unmarshal works as usual
```go
err := json.Unmarshal(data, &result)
```

Not cleaning fields UnmarshalNoZero looks similar
```go
err := json.UnmarshalNoZero(data, &result)
```

Perhaps you need only one subobject at document, easily and affectivly
```go
err := json.Wrap(data).Get("users", 0, "profile").Unmarshal(&result)
```

Or you need only one field
```go
city, err := json.Wrap(data).Get("users", 0, "profile", "location", "city").CheckString()
```

You may do the same for stream of objects (like application/json-seq)
```go
r := json.NewReader(stream)
for r.Type() != json.None {
    city, err := r.Get("profile", "location", "city").CheckString()
    if err != nil {
        break
    }
    _ = city
    r.GoOut(3) // read object to the end on 3 levels back (it's the depth we've got to by Get(...))
}
```

Or you can inspect an Array
```go
r.Get("profile", "languages")
for r.HasNext() {
    lang := r.NextString()
    _ = lang
}
```
or Object
```go
r.Get("profile", "health")
for r.HasNext() {
    key := r.NextString()
    switch string(key) {
    case "blood_pressure":
        _ = r.NextNumber()
    case "blood_type":
        _ = r.NextString()
    default:
        r.Skip() // don't forget to read unused values
    }
}
```

It's a single pass algorithm, but you can say it to remember position and return
```go
r.Lock()
val, err := r.Get("check","object","key").CheckString()
// ...
if ... {
    r.Unlock()
    return
}
r.Return()

err = r.Unmarshal(wanted_object)
// ...
```
Remember that all bytes that are locked are kept in memory, don't forget to unlock it when you are done

You can parse object in raw format without knowing structure (Full example at godoc under [Reader.Type() method](https://godoc.org/github.com/nikandfor/json#example-Reader-Type))
```go
parse = func(r *json.Reader, d int) {
    switch tp := r.Type(); tp {
    case json.String, json.Number, json.Bool, json.Null:
        // read by one of these accordingly
        //   r.CheckString()
        //   r.Int() or r.Float64()
        //   r.Bool()
        //   r.Skip() // you can skip any value like this
        val := r.NextBytes() // reads any value including object and array as raw bytes
        fmt.Printf("%*s is %v\n", d*4, val, tp)
    case json.ArrayStart:
        for r.HasNext() {
            parse(r, d+1)
        }
    case json.ObjStart:
        for r.HasNext() {
            key := r.NextString()
            fmt.Printf("%*s ->\n", d*4, key)
            parse(r, d+1)
        }
    default:
        err := r.Err()
        fmt.Printf("%*s got type %v err %#v\n", d*4, "", tp, err)
    }
}

data := `{"first": "string", "second": 123, "third": [1.1, 3.3, 7.7], "fourth": {"again": "string", "and": {"object": "here"}}}`
r := json.Wrap(data)

parse(r)

// Output:
// first ->
// "string" is String
// second ->
//      123 is Number
// third ->
//          1.1 is Number
//          3.3 is Number
//          7.7 is Number
// fourth ->
//    again ->
//     "string" is String
//      and ->
//       object ->
//           "here" is String
```
As you may noticed I didn't checked for errors. That's because I can do it only once at the end. Imagine we have typo in out previous example: first quote mark of `fourth` key is wrong.
```go
data := `{"first": "string", "second": 123, "third": [1.1, 3.3, 7.7], 'fourth": {"again": "string", "and": {"object": "here"}}}`
// ...
if err := r.Err(); err != nil {
    fmt.Printf("reader: %#v\n", err)
}
// Output:
// reader: parse error at pos 61: invalid character `.1, 3.3, 7.7], 'fourth": {"agai`
```
And we've got error message with position and context.
Errors could be printed an even better.
```go
// ...
if err := r.Err(); err != nil {
    fmt.Printf("reader: %+40v\n", err) // here we can choose extended format and context size (works for # either)
}
// Output:
// reader: parse error at pos 61: invalid character
// ...[1.1, 3.3, 7.7], 'fourth": {"again"...
// ____________________^____________________
```
Of course it works only with monospace fonts.

## Errors Checking
To tell the truth I must say that it doesn't tell you anything if you missed `,` or `:`, or some number you skipped was like this `45.67+13e44` (it does well if you parse numbers with r.Int() or r.Float64() or so methods, or unmarshal them).
All the errors that could be ignored are ignored. That's done for the sake of performance.

## Performance
Performance tests are the same as [jsoniter](https://github.com/json-iterator/go-benchmark/blob/master/src/github.com/json-iterator/go-benchmark/benchmark_medium_payload_test.go) and [json parser](https://github.com/buger/jsonparser/blob/master/benchmark/benchmark_medium_payload_test.go) have. And they are my closest competitors.

Medium size json object decoding
```
// for loop with single switch and for b[i] != '"' {i++} string skipper
BenchmarkRawLoopStructMedium-8                	 1000000	      1907 ns/op	1220.62 MB/s	       0 B/op	       0 allocs/op

// Unmarshal struct
BenchmarkDecodeStdStructMedium-8              	   50000	     31848 ns/op	  73.10 MB/s	    1960 B/op	      99 allocs/op
BenchmarkDecodeJsoniterStructMedium-8         	  200000	      6446 ns/op	 361.13 MB/s	     480 B/op	      45 allocs/op
BenchmarkDecodeEasyJsonMedium-8               	  200000	      7902 ns/op	 294.58 MB/s	     160 B/op	       4 allocs/op
BenchmarkDecodeNikandjsonStructMedium-8       	  200000	      6354 ns/op	 366.35 MB/s	      80 B/op	       2 allocs/op
// Buger don't have such function

// Get single key
BenchmarkDecodeBugerGetStructMedium-8         	  500000	      3495 ns/op	 665.93 MB/s	       0 B/op	       0 allocs/op
BenchmarkDecodeNikandjsonGetStructMedium-8    	  300000	      4475 ns/op	 520.12 MB/s	       0 B/op	       0 allocs/op
// others don't support such function
```
Large size json object decoding
```
// count array elements in large json
BenchmarkBugerJsonParserLarge-8    	   50000	     34488 ns/op	 815.43 MB/s	       0 B/op	       0 allocs/op
BenchmarkJsoniterLarge-8           	   20000	     94858 ns/op	 296.47 MB/s	   12976 B/op	    1133 allocs/op
BenchmarkNikandjsonGetLarge-8      	   30000	     52134 ns/op	 539.43 MB/s	       0 B/op	       0 allocs/op
```
As you can see my library is not the fastest, but it is as memory efficient as possible (as Buger as) and it has all the functions in single package.
Two allocs in `BenchmarkDecodeNikandjsonStructMedium` test is []byte to string conversion when struct fields are set (that could be avoided by `unsafe` hack, but in that situation it's not worth it).

There are two skipString function implementations: strict (default) that checks all strings to utf8 encoding correctness (it's still done for strings you get by NextString and CheckString methods) and fast that doesn't. It could be chosen at compilation time by setting build tag `unsafestrings`.
