// +build ignore

package json

import "fmt"

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
	err := WrapString(data).Search("person", "name").Unmarshal(&f)

	fmt.Printf("name: %+v, err: %v", f, err)

	// Output:
	// name: {GivenName:Leonid FamilyName:Bugaev}, err: <nil>
}
