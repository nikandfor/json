// Package json is an faster alternative to stdlib
//
// It allows you to parse bytes buffer or io.Reader stream with no allocs and
// without knowing document structure and field names.
// You also can do the same in opposite direction: building json in a streaming fashion.
//
// It's could be used as a replacement to "encoding/json". Or you can build
// complex analyzer on top of it.
//
// It's an one time reader, so read data are dropped if buffer gets full.
//
// It's doesn't do things you don't need. It won't parse or read anything until you ask.
// It doesn't decode escape sequences at strings you don't need.
// It copy readed string only if it contain escape sequences or if it doesn't fit
// info the buffer. And it's copied to reusable buffer.
// There are two versions on skipString functions: fast (default), that looks only
// for closing '"' byte, and strict, that check utf8 encoding also
// (you can choose between them by strict build tag).
// Reader.CompareKey doesn't check utf8 encoding and doesn't decode escape
// sequences (did you see object keys like this "\t\nkey \n"? if so, use
// (*Reader).NextString() instead).
//
// You should be very attentive since some checks were sacrificed in favor of
// speed. So you will not be notified if for example you forgot to read some
// value in the object and requested the next key.
//
// See examples for more ideas of usage
package json
