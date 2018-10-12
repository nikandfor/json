package json

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
)

func Marshal(v interface{}) ([]byte, error) {
	w := &Writer{}
	err := w.Marshal(v)
	return w.b, err
}

type Writer struct {
	b []byte
}

func (w *Writer) Marshal(v interface{}) error {
	if m, ok := v.(json.Marshaler); ok {
		data, err := m.MarshalJSON()
		if err != nil {
			return err
		}
		w.b = append(w.b, data...)
		return nil
	}
	if v == nil {
		w.b = append(w.b, "nil"...)
		return nil
	}

	switch v := v.(type) {
	case int, uint,
		int8, int16, int32, int64,
		uint8, uint16, uint32, uint64:
		w.b = append(w.b, []byte(fmt.Sprintf("%d", v))...)
		return nil
	case string:
		w.b = append(w.b, []byte(fmt.Sprintf("%q", v))...)
		return nil
	case *string:
		w.b = append(w.b, []byte(fmt.Sprintf("%q", *v))...)
		return nil
	case []byte:
		w.b = append(w.b, '"')
		w.b = append(w.b, []byte(base64.StdEncoding.EncodeToString(v))...)
		w.b = append(w.b, '"')
		return nil
	case *[]byte:
		w.b = append(w.b, '"')
		w.b = append(w.b, []byte(base64.StdEncoding.EncodeToString(*v))...)
		w.b = append(w.b, '"')
		return nil
	}

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			w.b = append(w.b, "nil"...)
			return nil
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Slice:
		w.b = append(w.b, '[')
		for i := 0; i < rv.Len(); i++ {
			if i != 0 {
				w.b = append(w.b, ',')
			}
			vi := rv.Index(i).Interface()
			err := w.Marshal(vi)
			if err != nil {
				return err
			}
		}
		w.b = append(w.b, ']')
	case reflect.Struct:
		w.b = append(w.b, '{')
		first := true
		for i := 0; i < rv.NumField(); i++ {
			name := getStructName(rv.Type(), i)
			if name == "" {
				continue
			}
			if !first {
				w.b = append(w.b, ',')
			} else {
				first = false
			}
			w.b = append(w.b, []byte(fmt.Sprintf("%q", name))...)
			w.b = append(w.b, ':')
			vi := rv.Field(i).Interface()
			err := w.Marshal(vi)
			if err != nil {
				return err
			}
		}
		w.b = append(w.b, '}')
	default:
		return w.Marshal(rv.Interface())
	}

	return nil
}
