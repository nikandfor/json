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
