// +build ignore

package json

import "testing"

func BenchmarkSearchUnmarshal(b *testing.B) {
	b.ReportAllocs()

	data := []byte(`{
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
}`)

	type FullName struct {
		GivenName  string
		FamilyName string
	}
	var f FullName
	var r Reader

	for i := 0; i < b.N; i++ {
		r.Reset(data)
		// Search only small needed subobject and unmarshal it
		r.Search("person", "name").Unmarshal(&f)
	}
}
