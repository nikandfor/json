package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	t.Skip("enable me")

	type B struct {
		E int
		D string
	}
	type A struct {
		I  int
		Ip *int
		S  string
		Sp *string
		B  []byte
		Bp *[]byte
		A  B
		Ap *B
	}

	exp := []byte(`{"I":1,"Ip":2,"S":"s_val","Sp":"sp_val","B":"AQID","Bp":"AwIB","A":{"E":4,"D":"d_val"},"Ap":{"E":6,"D":"ptr_d_val"}}`)

	var a A
	err := Unmarshal(exp, &a)
	if !assert.NoError(t, err) {
		return
	}

	data, err := Marshal(a)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, exp, data)
}

func TestMarshalTags(t *testing.T) {
	t.Skip("enable me")

	type B struct {
		E int    `json:"e"`
		D string `json:"d"`
	}
	type A struct {
		I  int     `json:"i"`
		Ip *int    `json:"ip"`
		S  string  `json:"s"`
		Sp *string `json:"sp"`
		B  []byte  `json:"b"`
		Bp *[]byte `json:"bp"`
		A  B       `json:"a"`
		Ap *B      `json:"ap"`
	}

	exp := []byte(`{"i":1,"ip":2,"s":"s_val","sp":"sp_val","b":"AQID","bp":"AwIB","a":{"e":4,"d":"d_val"},"ap":{"e":6,"d":"ptr_d_val"}}`)

	var a A
	err := Unmarshal(exp, &a)
	if !assert.NoError(t, err) {
		return
	}

	data, err := Marshal(a)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, exp, data)
}
