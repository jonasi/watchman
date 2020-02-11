package bser

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"reflect"
)

var (
	emptyValue = reflect.Value{}
)

// UnmarshalPDU unmarshal b into dest
func UnmarshalPDU(b []byte, dest interface{}) error {
	return NewDecoder(bytes.NewReader(b)).Decode(dest)
}

// UnmarshalValue unmarshal b into dest
func UnmarshalValue(b []byte, dest interface{}) error {
	return decodeValue(bytes.NewReader(b), reflect.ValueOf(dest), nil)
}

// NewDecoder returns an initialized Decoder
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r}
}

// A Decoder reads and decodes BSER values from an input stream
type Decoder struct {
	r io.Reader
}

// Decode reads the next BSER-encoded value from its
// input and stores it in the value pointed to by dest.
func (d *Decoder) Decode(dest interface{}) error {
	buf := make([]byte, 2)
	if _, err := d.r.Read(buf); err != nil {
		return err
	}

	if !bytes.Equal(buf, protocolPrefix) {
		return fmt.Errorf("Expected %x, found %x", protocolPrefix, buf)
	}

	var size int
	if err := decodeValue(d.r, reflect.ValueOf(&size), nil); err != nil {
		return err
	}

	buf = make([]byte, size)
	if _, err := d.r.Read(buf); err != nil {
		return err
	}

	return decodeValue(bytes.NewReader(buf), reflect.ValueOf(dest), nil)
}

func decodeValue(r io.Reader, dest reflect.Value, buf *[]byte) error {
	if dest != emptyValue && dest.Kind() != reflect.Ptr {
		return fmt.Errorf("Invalid dest passed in. Expected ptr, found: %s", dest.Kind())
	}

	var (
		u      Unmarshaler
		err    error
		offset = 0
	)

	if dest != emptyValue && dest.Type().Implements(typUnmarshaler) {
		u = dest.Interface().(Unmarshaler)
		if buf == nil {
			buf = &[]byte{}
			offset = 0
		} else {
			offset = len(*buf)
		}
		dest = emptyValue
	}

	if dest != emptyValue {
		dest = dest.Elem()
	}

	l := make([]byte, 1)
	if _, err := r.Read(l); err != nil {
		return err
	}

	if buf != nil {
		*buf = append(*buf, l...)
	}

	switch l[0] {
	case 0x00:
		err = decodeArray(r, dest, buf)
	case 0x01:
		err = decodeObject(r, dest, buf)
	case 0x02:
		err = decodeString(r, dest, buf)
	case 0x03:
		err = decodeInt8(r, dest, buf)
	case 0x04:
		err = decodeInt16(r, dest, buf)
	case 0x05:
		err = decodeInt32(r, dest, buf)
	case 0x06:
		err = decodeInt64(r, dest, buf)
	case 0x07:
		err = decodeReal(r, dest, buf)
	case 0x08:
		if dest != emptyValue {
			err = decodeBool(true, dest)
		}
	case 0x09:
		if dest != emptyValue {
			err = decodeBool(false, dest)
		}
	case 0x0a:
		if dest != emptyValue {
			dest.Set(reflect.ValueOf(nil))
		}
	case 0x0b:
		err = decodeTemplate(r, dest, buf)
	default:
		err = fmt.Errorf("Invalid type marker found: %x", l[0])
	}

	if err == nil && u != nil {
		err = u.UnmarshalBSER((*buf)[offset:])
	}

	return err
}

func decodeArray(r io.Reader, dest reflect.Value, buf *[]byte) error {
	dest, err := prep(dest, typGenericSlice)
	if err != nil {
		return err
	}

	if dest != emptyValue && dest.Kind() != reflect.Slice && dest.Kind() != reflect.Array {
		return fmt.Errorf("Expected slice or array, found %s", dest.Kind())
	}

	var length int
	if err = decodeValue(r, reflect.ValueOf(&length), buf); err != nil {
		return err
	}

	if dest != emptyValue && dest.Kind() == reflect.Slice {
		dest.Set(reflect.MakeSlice(dest.Type(), length, length))
	}

	for i := 0; i < length; i++ {
		if i >= dest.Len() {
			continue
		}
		v := emptyValue
		if dest != emptyValue {
			v = reflect.New(dest.Type().Elem())
		}

		if err := decodeValue(r, v, buf); err != nil {
			return err
		}

		if dest != emptyValue {
			dest.Index(i).Set(v.Elem())
		}
	}

	return nil
}

func decodeObject(r io.Reader, dest reflect.Value, buf *[]byte) error {
	dest, err := prep(dest, typGenericMap)
	if err != nil {
		return err
	}
	switch k := dest.Kind(); k {
	case reflect.Invalid:
		if buf == nil {
			panic("should only get reflect.Invalid if buf is non-nil")
		}
		var fields int
		if err := decodeValue(r, reflect.ValueOf(&fields), buf); err != nil {
			return err
		}
		for i := 0; i < fields; i++ {
			if err := decodeValue(r, emptyValue, buf); err != nil {
				return err
			}

			if err := decodeValue(r, emptyValue, buf); err != nil {
				return err
			}
		}
	case reflect.Map:
		if dest.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("Map must have key type of string, found: %s", dest.Type().Key())
		}

		var fields int
		if err := decodeValue(r, reflect.ValueOf(&fields), buf); err != nil {
			return err
		}

		for i := 0; i < fields; i++ {
			var field string
			fieldV := reflect.ValueOf(&field)
			if err := decodeValue(r, fieldV, buf); err != nil {
				return err
			}

			v := reflect.New(dest.Type().Elem())
			if err := decodeValue(r, v, buf); err != nil {
				return err
			}
			dest.SetMapIndex(fieldV.Elem(), v.Elem())
		}
	case reflect.Struct:
		tfields := fields(dest.Type())
		var fields int
		if err := decodeValue(r, reflect.ValueOf(&fields), buf); err != nil {
			return err
		}

		for i := 0; i < fields; i++ {
			var field string
			if err := decodeValue(r, reflect.ValueOf(&field), buf); err != nil {
				return err
			}

			f, ok := tfields.field(field)
			if !ok {
				return fmt.Errorf("Field %s not found", field)
			}

			if err := decodeValue(r, dest.FieldByIndex(f.Index).Addr(), buf); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("Expected struct or map, found %s", k)
	}
	return nil
}

func decodeString(r io.Reader, dest reflect.Value, buf *[]byte) error {
	dest, err := prep(dest, typString)
	if err != nil {
		return err
	}

	var size int
	if err := decodeValue(r, reflect.ValueOf(&size), buf); err != nil {
		return err
	}

	b := make([]byte, size)
	if _, err := r.Read(b); err != nil {
		return err
	}

	if dest != emptyValue {
		dest.SetString(string(b))
	}
	if buf != nil {
		*buf = append(*buf, b...)
	}
	return nil
}

func decodeInt8(r io.Reader, dest reflect.Value, buf *[]byte) error {
	dest, err := prep(dest, typInt8)
	if err != nil {
		return err
	}

	b := make([]byte, 1)
	if _, err := r.Read(b); err != nil {
		return err
	}

	if dest != emptyValue {
		dest.SetInt(int64(b[0]))
	}
	if buf != nil {
		*buf = append(*buf, b...)
	}
	return nil
}

func decodeInt16(r io.Reader, dest reflect.Value, buf *[]byte) error {
	dest, err := prep(dest, typInt16)
	if err != nil {
		return err
	}

	b := make([]byte, 2)
	if _, err := r.Read(b); err != nil {
		return err
	}

	if dest != emptyValue {
		v := order.Uint16(b)
		dest.SetInt(int64(v))
	}
	if buf != nil {
		*buf = append(*buf, b...)
	}
	return nil
}

func decodeInt32(r io.Reader, dest reflect.Value, buf *[]byte) error {
	dest, err := prep(dest, typInt32)
	if err != nil {
		return err
	}

	b := make([]byte, 4)
	if _, err := r.Read(b); err != nil {
		return err
	}

	if dest != emptyValue {
		v := order.Uint32(b)
		dest.SetInt(int64(v))
	}
	if buf != nil {
		*buf = append(*buf, b...)
	}
	return nil
}

func decodeInt64(r io.Reader, dest reflect.Value, buf *[]byte) error {
	dest, err := prep(dest, typInt64)
	if err != nil {
		return err
	}

	b := make([]byte, 8)
	if _, err := r.Read(b); err != nil {
		return err
	}

	if dest != emptyValue {
		v := order.Uint64(b)
		dest.SetInt(int64(v))
	}
	if buf != nil {
		*buf = append(*buf, b...)
	}
	return nil
}

func decodeReal(r io.Reader, dest reflect.Value, buf *[]byte) error {
	dest, err := prep(dest, typFloat64)
	if err != nil {
		return err
	}

	b := make([]byte, 8)
	if _, err := r.Read(b); err != nil {
		return err
	}

	if dest != emptyValue {
		v := order.Uint64(b)
		dest.SetFloat(math.Float64frombits(v))
	}
	if buf != nil {
		*buf = append(*buf, b...)
	}

	return nil
}

func decodeBool(v bool, dest reflect.Value) error {
	dest, err := prep(dest, typBool)
	if err != nil {
		return err
	}

	dest.SetBool(v)
	return nil
}

func decodeTemplate(r io.Reader, dest reflect.Value, buf *[]byte) error {
	dest, err := prep(dest, typGenericSlice)
	if err != nil {
		return err
	}

	if dest != emptyValue && dest.Kind() != reflect.Slice {
		return fmt.Errorf("Expected slice, found %s", dest.Kind())
	}

	var fieldNames []string
	if err := decodeValue(r, reflect.ValueOf(&fieldNames), buf); err != nil {
		return err
	}

	var length int
	if err := decodeValue(r, reflect.ValueOf(&length), buf); err != nil {
		return err
	}

	if dest != emptyValue {
		dest.Set(reflect.MakeSlice(dest.Type(), length, length))
	}

	var sfields structFields
	for i := 0; i < length; i++ {
		item := emptyValue
		if dest != emptyValue {
			item = dest.Index(i)
			item, err = prep(item, typGenericMap)
			if err != nil {
				return err
			}

			if item.Kind() != reflect.Map && item.Kind() != reflect.Struct {
				return fmt.Errorf("Expected slice of struct or slice of map, found slice of %s", item.Kind())
			}
		}

		for _, field := range fieldNames {
			switch item.Kind() {
			case reflect.Invalid:
				if buf == nil {
					panic("should only get reflect.Invalid if buf is non-nil")
				}
				v := emptyValue
				if err := decodeValue(r, v, buf); err != nil {
					return err
				}
			case reflect.Map:
				v := reflect.New(item.Type().Elem())
				if err := decodeValue(r, v, buf); err != nil {
					return err
				}
				item.SetMapIndex(reflect.ValueOf(field), v.Elem())
			case reflect.Struct:
				if sfields == nil {
					sfields = fields(item.Type())
				}

				f, ok := sfields.field(field)
				if !ok {
					return fmt.Errorf("Field %s not found", field)
				}

				if err := decodeValue(r, item.FieldByIndex(f.Index).Addr(), buf); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func prep(v reflect.Value, typ reflect.Type) (reflect.Value, error) {
	if v.Kind() != reflect.Interface {
		return v, nil
	}

	if v.Type().NumMethod() == 0 {
		val := reflect.New(typ)
		switch typ.Kind() {
		case reflect.Slice:
			val.Elem().Set(reflect.MakeSlice(typ, 0, 0))
		case reflect.Map:
			val.Elem().Set(reflect.MakeMap(typ))
		}

		v.Set(val)
		return val.Elem(), nil
	}

	return reflect.Value{}, fmt.Errorf("Interface found, but expected %s", typ)
}
