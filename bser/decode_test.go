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
