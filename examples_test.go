package json_test

import (
	"fmt"
	"strconv"

	"nikand.dev/go/json"
	"nikand.dev/go/json/benchmarks_data"
)

func ExampleIterator() {
	var d json.Iterator
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

func ExampleIterator_multipleValues() {
	var err error // to not to shadow i in a loop
	var d json.Iterator
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

func ExampleIterator_Seek_seekIter() {
	err := func(b []byte) error {
		var d json.Iterator

		i, err := d.Seek(b, 0, "topics", "topics")
		if err != nil {
			return fmt.Errorf("seek topics: %w", err)
		}

		i, err = d.Enter(b, i, json.Array)

		for err == nil && d.ForMore(b, &i, json.Array, &err) {
			var id int
			var title []byte

			i, err = d.IterFunc(b, i, json.Object, func(k, v []byte) error {
				switch string(k) {
				case "id":
					x, err := strconv.ParseInt(string(v), 10, 64)
					if err != nil {
						return fmt.Errorf("parse id: %w", err)
					}

					id = int(x)
				case "title":
					title, _, err = d.DecodeString(v, 0, title[:0])
					if err != nil {
						return fmt.Errorf("decode title: %w", err)
					}
				}

				return nil
			})

			fmt.Printf("> %3d %s\n", id, title)

		}

		if err != nil {
			return fmt.Errorf("iter topics: %w", err)
		}

		return nil
	}(benchmarks_data.LargeFixture)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	// Output:
	// >   8 Welcome to Metabase's Discussion Forum
	// > 169 Formatting Dates
	// > 168 Setting for google api key
	// > 167 Cannot see non-US timezones on the admin
	// > 164 External (Metabase level) linkages in data schema
	// > 155 Query working on "Questions" but not in "Pulses"
	// > 161 Pulses posted to Slack don't show question output
	// > 152 Should we build Kafka connecter or Kafka plugin
	// > 147 Change X and Y on graph
	// > 142 Issues sending mail via office365 relay
	// > 137 I see triplicates of my mongoDB collections
	// > 140 Google Analytics plugin
	// > 138 With-mongo-connection failed: bad connection details:
	// > 133 "We couldn't understand your question." when I query mongoDB
	// > 129 My bar charts are all thin
	// > 106 What is the expected return order of columns for graphing results when using raw SQL?
	// > 131 Set site url from admin panel
	// > 127 Internationalization (i18n)
	// > 109 Returning raw data with no filters always returns We couldn't understand your question
	// > 103 Support for Cassandra?
	// > 128 Mongo query with Date breaks [solved: Mongo 3.0 required]
	// >  23 Can this connect to MS SQL Server?
	// > 121 Cannot restart metabase in docker
	// >  85 Edit Max Rows Count
	// >  96 Creating charts by querying more than one table at a time
	// >  90 Trying to add RDS postgresql as the database fails silently
	// >  17 Deploy to Heroku isn't working
	// > 100 Can I use DATEPART() in SQL queries?
	// >  98 Feature Request: LDAP Authentication
	// >  87 Migrating from internal H2 to Postgres
}
