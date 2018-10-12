package json

import (
	"encoding/json"
	"reflect"
	"strings"
)

func (v *Value) Unmarshal(r interface{}) error {
	if u, ok := r.(json.Unmarshaler); ok {
		return u.UnmarshalJSON(v.buf)
	}
	rv := reflect.ValueOf(r)
	return v.unmarshal(rv)
}

func (v *Value) unmarshal(rv reflect.Value) error {
	for rv.Kind() == reflect.Ptr {
		ok, err := v.IsNull()
		if err == nil && ok {
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
		elt := rv.Type().Elem()

		res := reflect.MakeSlice(rv.Type(), 0, 0)

		i, err := v.ArrayIter()
		if err != nil {
			return err
		}
		for i.HasNext() {
			n := i.Next()
			re := reflect.New(elt)
			err := n.unmarshal(re)
			if err != nil {
				return err
			}
			res = reflect.Append(res, re.Elem())
		}

		if res.Len() != 0 {
			rv.Set(res)
		} else {
			rv.Set(reflect.Zero(rv.Type()))
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
			err := n.unmarshal(rv.Index(j))
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
		rv.Set(reflect.Zero(rv.Type()))

		i, err := v.ObjectIter()
		if err != nil {
			return err
		}

		for i.HasNext() {
			k, val := i.Next()

			fv, ok := getStructField(rv, k.MustCheckString())

			if !ok {
				continue
			}

			err = val.unmarshal(fv)
			if err != nil {
				return err
			}
		}

	default:
		panic(rv)
	}
	return nil
}

var (
	zeroVal           reflect.Value
	structFieldsCache map[reflect.Type]map[string]int
)

func getStructField(tv reflect.Value, f string) (v_ reflect.Value, ok_ bool) {
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
		return zeroVal, false
	}
	return tv.Field(fi), true
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
