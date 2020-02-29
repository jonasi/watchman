package bser

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

type nonDecodable interface {
	valid() bool
}

type decodeTest struct {
	encoded      []byte
	expectedData interface{}
	expectErr    bool
	doDecode     func(decoder *Decoder) (interface{}, error)
}

var decodeTests = map[string]decodeTest{
	"single_object": {
		encoded: []byte(
			"\x00\x01\x05\x19\x00\x00\x00\x01\x03\x02\x02\x03\x04Name\x02\x03\x04fred\x02\x03\x03Age\x03\x14",
		),
		expectedData: person{Name: "fred", Age: 20},
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := person{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"pos_int8_array": {
		encoded: []byte(
			"\x00\x01\x03\x0f\x00\x03\x06\x03\x01\x03\x02\x03\x03\x03\x04\x03\x05\x03\x06",
		),
		expectedData: [6]int8{1, 2, 3, 4, 5, 6},
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst [6]int8
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"int8_array_insufficient_size": {
		encoded: []byte(
			"\x00\x01\x03\x0f\x00\x03\x06\x03\x01\x03\x02\x03\x03\x03\x04\x03\x05\x03\x06",
		),
		expectedData: [3]int8{1, 2, 3},
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst [3]int8
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"int8_array_greater_size": {
		encoded: []byte(
			"\x00\x01\x03\x0f\x00\x03\x06\x03\x01\x03\x02\x03\x03\x03\x04\x03\x05\x03\x06",
		),
		expectedData: [8]int8{1, 2, 3, 4, 5, 6, 0, 0},
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst [8]int8
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"string_slice": {
		encoded: []byte(
			"\x00\x01\x03\x1b\x00\x03\x06\x02\x03\x01a\x02\x03\x01b\x02\x03\x01c\x02\x03\x01d\x02\x03\x01e\x02\x03\x01f",
		),
		expectedData: []string{"a", "b", "c", "d", "e", "f"},
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []string{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"pos_int8_slice": {
		encoded: []byte(
			"\x00\x01\x03\x0f\x00\x03\x06\x03\x01\x03\x02\x03\x03\x03\x04\x03\x05\x03\x06",
		),
		expectedData: []int8{1, 2, 3, 4, 5, 6},
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []int8{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"neg_int8_slice": {
		encoded: []byte(
			"\x00\x01\x03\x0f\x00\x03\x06\x03\xff\x03\xfe\x03\xfd\x03\xfc\x03\xfb\x03\xfa",
		),
		expectedData: []int8{-1, -2, -3, -4, -5, -6},
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []int8{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"pos_int16_slice": {
		encoded: []byte(
			"\x00\x01\x03\x15\x00\x03\x06\x04\xe9\x03\x04\xea\x03\x04\xeb\x03\x04\xec\x03\x04\xed\x03\x04\xee\x03",
		),
		expectedData: []int16{1001, 1002, 1003, 1004, 1005, 1006},
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []int16{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"neg_int16_slice": {
		encoded: []byte(
			"\x00\x01\x03\x15\x00\x03\x06\x04\x17\xfc\x04\x16\xfc\x04\x15\xfc\x04\x14\xfc\x04\x13\xfc\x04\x12\xfc",
		),
		expectedData: []int16{-1001, -1002, -1003, -1004, -1005, -1006},
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []int16{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"pos_int32_slice": {
		encoded: []byte(
			"\x00\x01\x03!\x00\x03\x06\x05\xa1\x86\x01\x00\x05\xa2\x86\x01\x00\x05\xa3\x86\x01\x00\x05\xa4\x86\x01\x00\x05\xa5\x86\x01\x00\x05\xa6\x86\x01\x00",
		),
		expectedData: []int32{100001, 100002, 100003, 100004, 100005, 100006},
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []int32{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"neg_int32_slice": {
		encoded: []byte(
			"\x00\x01\x03!\x00\x03\x06\x05_y\xfe\xff\x05^y\xfe\xff\x05]y\xfe\xff\x05\\y\xfe\xff\x05[y\xfe\xff\x05Zy\xfe\xff",
		),
		expectedData: []int32{-100001, -100002, -100003, -100004, -100005, -100006},
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []int32{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"pos_int64_slice": {
		encoded: []byte(
			"\x00\x01\x039\x00\x03\x06\x06\x00^\xd0\xb2\x00\x00\x00\x00\x06\x01^\xd0\xb2\x00\x00\x00\x00\x06\x02^\xd0\xb2\x00\x00\x00\x00\x06\x03^\xd0\xb2\x00\x00\x00\x00\x06\x04^\xd0\xb2\x00\x00\x00\x00\x06\x05^\xd0\xb2\x00\x00\x00\x00",
		),
		expectedData: []int64{3_000_000_000, 3_000_000_001, 3_000_000_002, 3_000_000_003, 3_000_000_004, 3_000_000_005},
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []int64{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"neg_int64_slice": {
		encoded: []byte(
			"\x00\x01\x039\x00\x03\x06\x06\x00\xa2/M\xff\xff\xff\xff\x06\xff\xa1/M\xff\xff\xff\xff\x06\xfe\xa1/M\xff\xff\xff\xff\x06\xfd\xa1/M\xff\xff\xff\xff\x06\xfc\xa1/M\xff\xff\xff\xff\x06\xfb\xa1/M\xff\xff\xff\xff",
		),
		expectedData: []int64{-3_000_000_000, -3_000_000_001, -3_000_000_002, -3_000_000_003, -3_000_000_004, -3_000_000_005},
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []int64{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"float64_slice": {
		encoded: []byte(
			"\x00\x01\x039\x00\x03\x06\x07{\x14\xaeG\xe1z\x84?\x07{\x14\xaeG\xe1z\x94?\x07\xb8\x1e\x85\xebQ\xb8\x9e?\x07{\x14\xaeG\xe1z\xa4?\x07\x9a\x99\x99\x99\x99\x99\xa9?\x07\xb8\x1e\x85\xebQ\xb8\xae?",
		),
		expectedData: []float64{0.01, 0.02, 0.03, 0.04, 0.05, 0.06},
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []float64{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"map_str_bool": {
		encoded: []byte(
			"\x00\x01\x05!\x00\x00\x00\x01\x03\x06\x02\x03\x01a\x08\x02\x03\x01b\t\x02\x03\x01c\x08\x02\x03\x01d\t\x02\x03\x01e\x08\x02\x03\x01f\t",
		),
		expectedData: map[string]bool{
			"a": true,
			"b": false,
			"c": true,
			"d": false,
			"e": true,
			"f": false,
		},
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := map[string]bool{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"template_arr": {
		// copied from https://facebook.github.io/watchman/docs/bser.html#array-of-templated-objects
		encoded: []byte{
			0x00, 0x01, 0x03, 0x28,
			0x0b,
			0x00,
			0x03, 0x02,
			0x02,
			0x03, 0x04,
			0x6e, 0x61, 0x6d, 0x65,
			0x02,
			0x03, 0x03,
			0x61, 0x67, 0x65,
			0x03, 0x03,
			0x02,
			0x03, 0x04,
			0x66, 0x72, 0x65, 0x64,
			0x03, 0x14,
			0x02,
			0x03, 0x04,
			0x70, 0x65, 0x74, 0x65,
			0x03, 0x1e,
			0x0c,
			0x03, 0x19,
		},
		expectedData: []person{
			{Name: "fred", Age: 20},
			{Name: "pete", Age: 30},
			{Age: 25},
		},
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []person{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"string_alias": {
		encoded: []byte(
			"\x00\x01\x03\x08\x02\x03\x05hello",
		),
		expectedData: stringAlias("hello"),
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst stringAlias
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"int8_alias": {
		encoded: []byte(
			"\x00\x01\x03\x02\x03\x02",
		),
		expectedData: int8Alias(2),
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst int8Alias
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"int16_alias": {
		encoded: []byte(
			"\x00\x01\x03\x03\x04\xe9\x03",
		),
		expectedData: int16Alias(1001),
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst int16Alias
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"int32_alias": {
		encoded: []byte(
			"\x00\x01\x03\x05\x05\xa1\x86\x01\x00",
		),
		expectedData: int32Alias(100001),
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst int32Alias
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"int64_alias": {
		encoded: []byte(
			"\x00\x01\x03\x09\x06\x01^\xd0\xb2\x00\x00\x00\x00",
		),
		expectedData: int64Alias(3_000_000_001),
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst int64Alias
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"int_alias_int32_value": {
		encoded: []byte(
			"\x00\x01\x03\x05\x05\xa1\x86\x01\x00",
		),
		expectedData: int64Alias(100001),
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst int64Alias
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"bool_alias": {
		encoded: []byte(
			"\x00\x01\x03\x01\x09",
		),
		expectedData: boolAlias(false),
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst boolAlias
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"invalid_size_type_identifier": {
		encoded: []byte(
			"\x00\x01\xff\x02\x03\x10", // encoded 16 with int8 type in header replaced with 0xff
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst int8
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"invalid_type_identifier": {
		encoded: []byte(
			"\x00\x01\x03\x02\xff\x10", // encoded 16 with int8 type replaced with 0xff
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst int8
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"slice_of_invalid_type": {
		encoded: []byte(
			"\x00\x01\x03\x07\x00\x03\x02\xff\x01\xff\x02", // encoded [1, 2] with int8 type replaced with 0xff
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst = []int8{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"slice_with_invalid_length_type": {
		encoded: []byte(
			"\x00\x01\x03\x07\x00\xff\x02\x03\x01\x03\x02", // encoded [1, 2] with int8 length type replaced with 0xff
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst = []int8{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"object_with_invalid_key_type": {
		encoded: []byte(
			"\x00\x01\x05\x19\x00\x00\x00\x01\x03\x02\xff\x03\x04Name\x02\x03\x04fred\x02\x03\x03Age\x03\x14",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := person{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"object_with_invalid_value_type": {
		encoded: []byte(
			"\x00\x01\x05\x19\x00\x00\x00\x01\x03\x02\x02\x03\x04Name\x02\x03\x04fred\x02\x03\x03Age\xff\x14",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := person{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"map_with_invalid_key_type": {
		encoded: []byte(
			"\x00\x01\x03\x08\x01\x03\x01\xff\x03\x01a\x03\x01", // encoded {"a": 1} with string type replaced with 0xff
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := map[string]int{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"map_with_invalid_value_type": {
		encoded: []byte(
			"\x00\x01\x03\x08\x01\x03\x01\x02\x03\x01a\xff\x01", // encoded {"a": 1} with int8 type replaced with 0xff
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := map[string]int{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"non_pointer_dst": {
		encoded: []byte(
			"\x00\x01\x05\x19\x00\x00\x00\x01\x03\x02\x02\x03\x04Name\x02\x03\x04fred\x02\x03\x03Age\x03\x14",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := person{}
			err := decoder.Decode(dst)
			return dst, err
		},
	},
	"int8_array_to_object": {
		encoded: []byte(
			"\x00\x01\x03\x0f\x00\x03\x06\x03\x01\x03\x02\x03\x03\x03\x04\x03\x05\x03\x06",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := person{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"int8_array_to_string_arr": {
		encoded: []byte(
			"\x00\x01\x03\x0f\x00\x03\x06\x03\x01\x03\x02\x03\x03\x03\x04\x03\x05\x03\x06",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []string{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"int8_array_to_interface": {
		encoded: []byte(
			"\x00\x01\x03\x0f\x00\x03\x06\x03\x01\x03\x02\x03\x03\x03\x04\x03\x05\x03\x06",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst nonDecodable
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"int8_array_to_interface_arr": {
		encoded: []byte(
			"\x00\x01\x03\x0f\x00\x03\x06\x03\x01\x03\x02\x03\x03\x03\x04\x03\x05\x03\x06",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst []nonDecodable
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"int16_arr_to_string_arr": {
		encoded: []byte(
			"\x00\x01\x03\x15\x00\x03\x06\x04\xe9\x03\x04\xea\x03\x04\xeb\x03\x04\xec\x03\x04\xed\x03\x04\xee\x03",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []string{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"int16_arr_to_interface_arr": {
		encoded: []byte(
			"\x00\x01\x03\x15\x00\x03\x06\x04\xe9\x03\x04\xea\x03\x04\xeb\x03\x04\xec\x03\x04\xed\x03\x04\xee\x03",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst []nonDecodable
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"int32_arr_to_string_arr": {
		encoded: []byte(
			"\x00\x01\x03!\x00\x03\x06\x05\xa1\x86\x01\x00\x05\xa2\x86\x01\x00\x05\xa3\x86\x01\x00\x05\xa4\x86\x01\x00\x05\xa5\x86\x01\x00\x05\xa6\x86\x01\x00",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []string{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"int32_arr_to_interface_arr": {
		encoded: []byte(
			"\x00\x01\x03!\x00\x03\x06\x05\xa1\x86\x01\x00\x05\xa2\x86\x01\x00\x05\xa3\x86\x01\x00\x05\xa4\x86\x01\x00\x05\xa5\x86\x01\x00\x05\xa6\x86\x01\x00",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst []nonDecodable
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"int64_arr_to_string_arr": {
		encoded: []byte(
			"\x00\x01\x039\x00\x03\x06\x06\x00^\xd0\xb2\x00\x00\x00\x00\x06\x01^\xd0\xb2\x00\x00\x00\x00\x06\x02^\xd0\xb2\x00\x00\x00\x00\x06\x03^\xd0\xb2\x00\x00\x00\x00\x06\x04^\xd0\xb2\x00\x00\x00\x00\x06\x05^\xd0\xb2\x00\x00\x00\x00",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []string{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"int64_arr_to_interface_arr": {
		encoded: []byte(
			"\x00\x01\x039\x00\x03\x06\x06\x00^\xd0\xb2\x00\x00\x00\x00\x06\x01^\xd0\xb2\x00\x00\x00\x00\x06\x02^\xd0\xb2\x00\x00\x00\x00\x06\x03^\xd0\xb2\x00\x00\x00\x00\x06\x04^\xd0\xb2\x00\x00\x00\x00\x06\x05^\xd0\xb2\x00\x00\x00\x00",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst []nonDecodable
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"object_to_int_slice": {
		encoded: []byte(
			"\x00\x01\x05\x19\x00\x00\x00\x01\x03\x02\x02\x03\x04Name\x02\x03\x04fred\x02\x03\x03Age\x03\x14",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []int{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"string_slice_to_int_array": {
		encoded: []byte(
			"\x00\x01\x03\x1b\x00\x03\x06\x02\x03\x01a\x02\x03\x01b\x02\x03\x01c\x02\x03\x01d\x02\x03\x01e\x02\x03\x01f",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []int{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"dst_map_has_int_keys": {
		encoded: []byte(
			"\x00\x01\x05\x0b\x00\x00\x00\x01\x03\x01\x02\x03\x011\x02\x03\x01a",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := map[int]string{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"map_to_interface": {
		encoded: []byte(
			"\x00\x01\x05\x0b\x00\x00\x00\x01\x03\x01\x02\x03\x011\x02\x03\x01a",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst nonDecodable
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"bool_to_interface": {
		encoded: []byte(
			"\x00\x01\x05\x01\x00\x00\x00\x08",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst nonDecodable
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"bool_to_int": {
		encoded: []byte(
			"\x00\x01\x05\x01\x00\x00\x00\x08",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst int
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"float_to_string": {
		encoded: []byte(
			"\x00\x01\x05\t\x00\x00\x00\x07\xaeG\xe1z\x14\xae\xf3?",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst string
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"dst_map_has_incorrect_field_type": {
		encoded: []byte(
			"\x00\x01\x05\x0b\x00\x00\x00\x01\x03\x01\x02\x03\x011\x02\x03\x01a",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := map[string]int{} // encoded data was of type map[string]string
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"dst_object_has_incorrect_field_type": {
		encoded: []byte(
			"\x00\x01\x05\x19\x00\x00\x00\x01\x03\x02\x02\x03\x04Name\x02\x03\x04fred\x02\x03\x03Age\x03\x14",
		),
		expectErr: true,
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := struct {
				Name int // encoded object specifies this field should have string value
				Age  int
			}{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"struct_field": {
		encoded: []byte(
			"\x00\x01\x05*\x00\x00\x00\x01\x03\x03\x02\x03\x04Name\x02\x03\x04fred\x02\x03\x03Age\x03\x0c\x02\x03\x05Power\x02\x03\x06eating",
		),
		expectedData: superperson{Person: Person{Name: "fred", Age: 12}, Power: "eating"},
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst superperson
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"ptr_struct_field": {
		encoded: []byte(
			"\x00\x01\x05*\x00\x00\x00\x01\x03\x03\x02\x03\x04Name\x02\x03\x04fred\x02\x03\x03Age\x03\x0c\x02\x03\x05Power\x02\x03\x06eating",
		),
		expectedData: superperson2{Person: &Person{Name: "fred", Age: 12}, Power: "eating"},
		doDecode: func(decoder *Decoder) (interface{}, error) {
			var dst superperson2
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
}

func TestDecode(t *testing.T) {
	for testName, testCase := range decodeTests {
		t.Run(testName, func(t *testing.T) {
			testDecode(t, testCase)
		})
	}
}

func testDecode(t *testing.T, testCase decodeTest) {
	t.Helper()
	var (
		buf     = bytes.NewBuffer(testCase.encoded)
		decoder = NewDecoder(buf)
	)

	dst, err := testCase.doDecode(decoder)
	if testCase.expectErr && err != nil {
		return
	}
	if testCase.expectErr && err == nil {
		t.Fatal("unexpectedly no error")
	}
	if !testCase.expectErr && err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !reflect.DeepEqual(dst, testCase.expectedData) {
		t.Fatalf("unexpected decoded dst:\n\nexpected = %#v\n\nactual = %#v", testCase.expectedData, dst)
	}
}

type decodeBench struct {
	encoded  []byte
	doDecode func(decoder *Decoder) (interface{}, error)
}

var decodeBenches = map[string]decodeBench{
	"single_object": {
		encoded: []byte(
			"\x00\x01\x05\x19\x00\x00\x00\x01\x03\x02\x02\x03\x04Name\x02\x03\x04fred\x02\x03\x03Age\x03\x14",
		),
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := person{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"string_slice": {
		encoded: []byte(
			"\x00\x01\x03\x1b\x00\x03\x06\x02\x03\x01a\x02\x03\x01b\x02\x03\x01c\x02\x03\x01d\x02\x03\x01e\x02\x03\x01f",
		),
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []string{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"pos_int8_slice": {
		encoded: []byte(
			"\x00\x01\x03\x0f\x00\x03\x06\x03\x01\x03\x02\x03\x03\x03\x04\x03\x05\x03\x06",
		),
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []int8{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"pos_int16_slice": {
		encoded: []byte(
			"\x00\x01\x03\x15\x00\x03\x06\x04\xe9\x03\x04\xea\x03\x04\xeb\x03\x04\xec\x03\x04\xed\x03\x04\xee\x03",
		),
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []int16{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"pos_int32_slice": {
		encoded: []byte(
			"\x00\x01\x03!\x00\x03\x06\x05\xa1\x86\x01\x00\x05\xa2\x86\x01\x00\x05\xa3\x86\x01\x00\x05\xa4\x86\x01\x00\x05\xa5\x86\x01\x00\x05\xa6\x86\x01\x00",
		),
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []int32{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"pos_int64_slice": {
		encoded: []byte(
			"\x00\x01\x039\x00\x03\x06\x06\x00^\xd0\xb2\x00\x00\x00\x00\x06\x01^\xd0\xb2\x00\x00\x00\x00\x06\x02^\xd0\xb2\x00\x00\x00\x00\x06\x03^\xd0\xb2\x00\x00\x00\x00\x06\x04^\xd0\xb2\x00\x00\x00\x00\x06\x05^\xd0\xb2\x00\x00\x00\x00",
		),
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []int64{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"float64_slice": {
		encoded: []byte(
			"\x00\x01\x039\x00\x03\x06\x07{\x14\xaeG\xe1z\x84?\x07{\x14\xaeG\xe1z\x94?\x07\xb8\x1e\x85\xebQ\xb8\x9e?\x07{\x14\xaeG\xe1z\xa4?\x07\x9a\x99\x99\x99\x99\x99\xa9?\x07\xb8\x1e\x85\xebQ\xb8\xae?",
		),
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := []float64{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
	"map_str_bool": {
		encoded: []byte(
			"\x00\x01\x05!\x00\x00\x00\x01\x03\x06\x02\x03\x01a\x08\x02\x03\x01b\t\x02\x03\x01c\x08\x02\x03\x01d\t\x02\x03\x01e\x08\x02\x03\x01f\t",
		),
		doDecode: func(decoder *Decoder) (interface{}, error) {
			dst := map[string]bool{}
			err := decoder.Decode(&dst)
			return dst, err
		},
	},
}

var benchDecErr error

func BenchmarkDecode(b *testing.B) {
	for benchName, benchCase := range decodeBenches {
		var err error
		b.Run(fmt.Sprintf("data=%s", benchName), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var (
					buf     = bytes.NewBuffer(benchCase.encoded)
					decoder = NewDecoder(buf)
				)
				_, err = benchCase.doDecode(decoder)
			}
		})
		benchDecErr = err
	}
}
