package json

import (
	"fmt"
	"io"
	"strings"
)

// This example shows how you can collect some info out of json without fully parsing it
func ExampleReader() {
	data := `[{"user":"Kyle","visits":134},{"user":"Rose","visits":45,"online":true},{"user":"Adam","visits":37}]`
	r := WrapString(data)

	var visits int
	for r.HasNext() {
		r.Get("visits")
		visits += r.MustInt()
		r.GoOut(1) // number of keys passed to Get()
	}

	fmt.Printf("Total visits: %d\n", visits)

	// Output:
	// Total visits: 216
}

func ExampleReader_Unmarshal() {
	data := `{
  "person": {
    "id": "d50887ca-a6ce-4e59-b89f-14f0b5d03b03",
    "name": {
      "fullName": "Leonid Bugaev",
      "givenName": "Leonid",
      "familyName": "Bugaev"
    },
    "email": "leonsbox@gmail.com",
    "gender": "male",
    "location": "Saint Petersburg, Saint Petersburg, RU",
  }
}`

	type FullName struct {
		GivenName  string
		FamilyName string
	}

	var f FullName

	// Get only small needed subobject and unmarshal it
	err := WrapString(data).Get("person", "name").Unmarshal(&f)

	fmt.Printf("name: %+v, err: %v", f, err)

	// Output:
	// name: {GivenName:Leonid FamilyName:Bugaev}, err: <nil>
}

func ExampleReader_messenger() {
	data := `
		{"status":"online","term":3,"messages":[
			{"text":"Tomorrow at 6pm Remember!!","sender":"Violet","term":3}]}
		{"status":"online","term":3,"messages":[
			{"text":"Tomorrow at 6pm Remember!!","sender":"Violet","term":3}]}
		{"status":"online","term":4,"messages":[
			{"text":"Tomorrow at 6pm Remember!!","sender":"Violet","term":3},
			{"text":"Hi, Nik! You won't believe who have I met just now!","sender":"Dan","term":4},
			{"text":"When did you see Dan last time?","sender":"Jane","term":4},
		]}
	`
	// Let's imagine that `stream` is a socket and we are taking server updates once a 5 seconds
	stream := strings.NewReader(data)

	type Message struct {
		Text   string
		Sender string
		Term   int
	}

	r := NewReader(stream)

	messagesTotal := 0
	lastTerm := 0
	for r.Type() != None {
		ok := r.Inspect(func(r *Reader) bool {
			term, err := r.Int()
			if err != nil {
				return false
			}
			res := term != lastTerm
			lastTerm = term
			return res
		}, "term")

		if !ok {
			continue
		}

		r.Inspect(func(r *Reader) bool {
			for r.HasNext() {
				var m Message
				err := r.Unmarshal(&m)
				if err != nil {
					return false
				}

				if m.Term != lastTerm {
					continue
				}

				messagesTotal++

				fmt.Printf("Message from %v: \t%v\n", m.Sender, m.Text)
			}
			return false
		}, "messages")
	}
	if err := r.Err(); err != io.EOF && err != nil {
		// process eny errors at one place
	}

	fmt.Printf("Connection closed. Total messages got: %d", messagesTotal)

	// Output:
	// Message from Violet: 	Tomorrow at 6pm Remember!!
	// Message from Dan: 	Hi, Nik! You won't believe who have I met just now!
	// Message from Jane: 	When did you see Dan last time?
	// Connection closed. Total messages got: 3
}

func ExampleReader_HasNext() {
	data := `{"a": [{"b": "c"}, {"d": "e"}],"f": true}`
	r := WrapString(data)

	r.Get("a")

	for r.HasNext() { // over array
		for r.HasNext() { // over array elements key-value pairs
			// we must use for loop here even if we know it's the only pair inside
			// because we have to read closing braket
			key := r.NextString()
			val := r.NextString()
			fmt.Printf("pair:  %q -> %q\n", key, val)
		}
	}

	fKey := r.NextString()
	switch r.Type() {
	case Bool:
		fVal, _ := r.Bool() // error could be here if there is truu instead of true for example
		fmt.Printf("%q -> %v\n", fKey, fVal)
	}

	r.GoOut(1) // to get out of the most outer object
	// it's the pair call to the first Get and we always have to call it if we want to read following data correctly

	fmt.Printf("end of buffer, next value type: %v\n", r.Type())

	// Output:
	// pair:  "b" -> "c"
	// pair:  "d" -> "e"
	// "f" -> true
	// end of buffer, next value type: None
}

func ExampleReader_Get() {
	data := `
	{"day": "Mon", "stats": {"views": {"by_partofday": [1, 2, 3]}}}
	{"day": "Tue", "stats": {"views": {"by_partofday": [3, 2, 3]}}}
	{"day": "Wed", "stats": {"views": {"by_partofday": [4, 1, 5]}}}
	{"day": "Thu", "stats": {"views": {"by_partofday": [1, 2, 2]}}}
	`

	s := strings.NewReader(data)

	r := NewReader(s)

	sum := 0
	days := 0
	for r.Type() != None {
		days++
		r.Get("stats", "views", "by_partofday") // goes inside to requested value
		for r.HasNext() {
			sum += r.MustInt()
		}
		r.GoOut(3) // goes back up at 3 levels (to the end of the current day object)
	}

	fmt.Printf("total views for %v days: %v\n", days, sum)

	// Output:
	// total views for 4 days: 29
}

func ExampleReader_Type() {
	data := `{"first": "string", "second": 123, "third": [1.1, 3.3, 7.7], "fourth": {"again": "string", "and": {"object": "here"}}}`
	var parse func(*Reader, int) error
	parse = func(r *Reader, d int) error {
		switch tp := r.Type(); tp {
		case String, Number, Bool, Null:
			// read by one of accordingly
			//   r.CheckString()
			//   r.Int() or r.Float64()
			//   r.Bool()
			//   r.Skip() // you can skip any value like this
			val := r.NextBytes() // reads any value including object and array as raw bytes
			fmt.Printf("%*s is %v\n", d*4, val, tp)
		case ArrayStart:
			for r.HasNext() {
				err := parse(r, d+1)
				if err != nil {
					return err
				}
			}
		case ObjStart:
			for r.HasNext() {
				key := r.NextString()
				fmt.Printf("%*s ->\n", d*4, key)
				err := parse(r, d+1)
				if err != nil {
					return err
				}
			}
		default:
			err := r.Err()
			fmt.Printf("%*s got type %v err %#v\n", d*4, "", tp, err)
		}
		return nil
	}

	r := WrapString(data)

	err := parse(r, 1)
	if err != nil {
		fmt.Printf("error: %v", err)
	}
	if err = r.Err(); err != nil {
		fmt.Printf("reader error: %v", err)
	}

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
}
