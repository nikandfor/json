package json

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"
)

func UnmarshalV0(data []byte, r interface{}) error {
	return Wrap(data).UnmarshalV0(r)
}

func (v *Value) UnmarshalV0(r interface{}) error {
	if u, ok := r.(json.Unmarshaler); ok {
		return u.UnmarshalJSON(v.buf)
	}
	rv := reflect.ValueOf(r)
	return v.unmarshalV0(rv)
}

func (v *Value) unmarshalV0(rv reflect.Value) error {
	for rv.Kind() == reflect.Ptr {
		ok, err := v.IsNull()
		isNil := err == nil && ok

		if isNil && rv.IsNil() {
			return nil
		}

		if isNil {
			rv = rv.Elem()
			rv.Set(reflect.Zero(rv.Type()))
			return nil
		}

		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Uint,
		reflect.Uint64,
		reflect.Uint32,
		reflect.Uint16,
		reflect.Uint8:
		r, err := v.Uint64()
		if err != nil {
			return err
		}
		rv.SetUint(r)
	case reflect.Int,
		reflect.Int64,
		reflect.Int32,
		reflect.Int16,
		reflect.Int8:
		r, err := v.Int64()
		if err != nil {
			return err
		}
		rv.SetInt(int64(r))
	case reflect.String:
		r, err := v.CheckString()
		if err != nil {
			return err
		}
		rv.SetString(r)
	case reflect.Slice:
		if tp, err := v.Type(); err == nil && tp == String && rv.Type().Elem().Kind() == reflect.Uint8 {
			str := v.Bytes()
			n := base64.StdEncoding.DecodedLen(len(str))
			if n < rv.Cap() {
				rv.Set(rv.Slice(0, n))
			} else {
				rv.Set(rv.Slice(0, rv.Cap()))
				for rv.Len() < n {
					rv.Set(reflect.Append(rv, reflect.Zero(rv.Type().Elem())))
				}
			}
			_, err := base64.StdEncoding.Decode(rv.Interface().([]byte), str)
			if err != nil {
				return err
			}
			return nil
		}

		i, err := v.ArrayIter()
		if err != nil {
			return err
		}
		j := 0
		for i.HasNext() {
			n := i.Next()

			if j < rv.Len() {
				err = n.unmarshalV0(rv.Index(j))
			} else if j < rv.Cap() {
				rv.Set(rv.Slice(0, j+1))
				err = n.unmarshalV0(rv.Index(j))
			} else {
				rv.Set(reflect.Append(rv, reflect.Zero(rv.Type().Elem())))
				err = n.unmarshalV0(rv.Index(j))
			}
			if err != nil {
				return err
			}
			j++
		}

		if j < rv.Len() {
			rv.Set(rv.Slice(0, j))
		}

		if rv.IsNil() {
			rv.Set(reflect.MakeSlice(rv.Type(), 0, 0))
		}

	case reflect.Array:
		elt := rv.Type().Elem()

		j := 0
		i, err := v.ArrayIter()
		if err != nil {
			return err
		}
		for i.HasNext() {
			n := i.Next()
			err := n.unmarshalV0(rv.Index(j))
			if err != nil {
				return err
			}
			j++
		}

		for j < rv.Len() {
			rv.Index(j).Set(reflect.Zero(elt))
			j++
		}

	case reflect.Struct:
		//	rv.Set(reflect.Zero(rv.Type()))
		vis := make([]bool, rv.NumField())

		i, err := v.ObjectIter()
		if err != nil {
			return err
		}

		for i.HasNext() {
			k, val := i.Next()

			fi, ok := getStructField(rv, k.MustCheckString())

			if !ok {
				continue
			}

			err = val.unmarshalV0(rv.Field(fi))
			if err != nil {
				return err
			}

			vis[fi] = true
		}

		for i, vis := range vis {
			if !vis {
				rv.Field(i).Set(reflect.Zero(rv.Field(i).Type()))
			}
		}

	default:
		panic(rv)
	}
	return nil
}

var (
	structFieldsCache     map[reflect.Type]map[string]int
	structFieldNamesCache map[reflect.Type][]string
)

func getStructField(tv reflect.Value, f string) (v_ int, ok_ bool) {
	t := tv.Type()

	if structFieldsCache == nil {
		structFieldsCache = make(map[reflect.Type]map[string]int)
	}
	sub, ok := structFieldsCache[t]
	if !ok {
		sub = buildFieldsCache(t)
		structFieldsCache[t] = sub
	}

	fi, ok := sub[f]
	if !ok {
		fi, ok = sub[strings.Title(f)]
	}
	if !ok {
		return 0, false
	}
	return fi, true
}

func getStructName(t reflect.Type, i int) string {
	if structFieldNamesCache == nil {
		structFieldNamesCache = make(map[reflect.Type][]string)
	}
	sub, ok := structFieldNamesCache[t]
	if !ok {
		sub = buildFieldNamesCache(t)
		structFieldNamesCache[t] = sub
	}

	return sub[i]
}

func buildFieldsCache(t reflect.Type) map[string]int {
	r := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		name := ft.Name
		tag, ok := ft.Tag.Lookup("json")
		if ok {
			if tag == "-" {
				continue
			}
			tags := strings.Split(tag, ",")
			name = tags[0]
		}
		if _, ok := r[name]; ok || name == "" {
			continue
		}
		r[name] = ft.Index[0]
	}
	return r
}

func buildFieldNamesCache(t reflect.Type) []string {
	r := make([]string, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		name := ft.Name
		tag, ok := ft.Tag.Lookup("json")
		if ok {
			if tag == "-" {
				continue
			}
			tags := strings.Split(tag, ",")
			name = tags[0]
		}
		r[i] = name
	}
	return r
}
