package bser

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

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
