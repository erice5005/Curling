package curling

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type SampleStruct struct {
	Name string `json:"name"`
}

func Test_ReaderForType(t *testing.T) {

	t.Parallel()

	tests := []struct {
		name        string
		dataset     interface{}
		datasetType reflect.Kind
		expectErr   bool
	}{
		{
			name: "Test with Struct",
			dataset: SampleStruct{
				Name: "test_struct",
			},
			expectErr:   false,
			datasetType: reflect.Struct,
		},
		{
			name:        "Test with String",
			dataset:     "test string",
			expectErr:   false,
			datasetType: reflect.String,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getReaderForType(tt.dataset)

			// Expected err, didn't get one
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected err, didn't get one: %w", err)
					t.Fail()
				} else {
					// expected err, got err, end successfully
					return
				}
			} else {
				if err != nil {
					t.Errorf("got err when none expected: %w", err)
					t.Fail()
				}
			}
			// Didn't expect error, got one

			buf := new(strings.Builder)
			_, err = io.Copy(buf, got)

			if err != nil {
				t.Fail()
			}

			passed := false
			switch tt.datasetType {
			case reflect.String:

				if assert.Equal(t, tt.dataset.(string), buf.String()) {
					passed = true
				}
			case reflect.Struct:
				marshedExpected, err := json.Marshal(tt.dataset)
				if err != nil {
					t.Fail()
				}
				if assert.JSONEq(t, string(marshedExpected), buf.String(), "should be equivalent") {
					passed = true
				}
			}

			if !passed {
				t.Fail()
			}
		})
	}
}

func Test_Request(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expected  interface{}
		expectErr bool
	}{
		{
			name:      "Test with String",
			expected:  "test string",
			expectErr: false,
		},
		// {
		// 	name: "Test with Struct",
		// 	expected: SampleStruct{
		// 		Name: "struct",
		// 	},
		// 	expectErr: false,
		// },
	}

	for _, tt := range tests {
		expected := tt.expected
		t.Run(tt.name, func(t *testing.T) {
			svr := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				fmt.Fprint(rw, expected)
			}))
			defer svr.Close()

			got, _, err := New(GET, svr.URL, nil).Do(tt.expected)
			if err != nil {
				t.Errorf("error running request: %w", err)
				t.Fail()
			}

			if string(got) != tt.expected {
				t.Errorf("expected: %v, got: %v", tt.expected, string(got))
				t.Fail()
			}
		})
	}

}

func BenchmarkSingleGet(b *testing.B) {
	svr := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprint(rw, nil)
	}))
	defer svr.Close()
	for n := 0; n < b.N; n++ {
		_, _, err := New(GET, svr.URL, nil).Do(nil)
		if err != nil {
			panic(err)
		}
	}
}
