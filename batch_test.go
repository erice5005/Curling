package curling

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func BenchmarkSingleBatchGet(b *testing.B) {
	svr := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprint(rw, nil)
	}))
	defer svr.Close()

	for n := 0; n < b.N; n++ {
		br := NewBatchRequest([]*BatchItem{
			{
				TargetURL:  svr.URL,
				Iterations: 4,
				Method:     GET,
				Delay:      1 * time.Nanosecond,
			},
		}, true)
		br.RunBatch()
	}
}

func BenchmarkMultiBatchGet(b *testing.B) {
	svr := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprint(rw, nil)
	}))
	defer svr.Close()

	for n := 0; n < b.N; n++ {
		br := NewBatchRequest([]*BatchItem{
			{
				TargetURL:  svr.URL,
				Iterations: 4,
				Method:     GET,
				Delay:      1 * time.Nanosecond,
			},
			{
				TargetURL:  svr.URL,
				Iterations: 6,
				Method:     GET,
				Delay:      2 * time.Nanosecond,
			},
			{
				TargetURL:  svr.URL,
				Iterations: 3,
				Method:     GET,
				Delay:      250 * time.Millisecond,
			},
		}, true)
		br.RunBatch()
	}
}

func Test_IndividualChannel(t *testing.T) {
	t.Parallel()
	svr := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprint(rw, r.Method)
	}))
	defer svr.Close()
	tests := []struct {
		name         string
		rx           []*BatchItem
		expectedIds  []string
		expectFail   bool
		expectGlobal bool
	}{
		{
			name: "run two different methods",
			rx: []*BatchItem{
				{
					Method:     GET,
					Iterations: 4,
					TargetURL:  svr.URL,
					Output: &ItemOutput{
						Id:     "GET",
						Output: make(chan OutputFrame),
					},
				},
				{
					Method:     POST,
					Iterations: 4,
					TargetURL:  svr.URL,
					Output: &ItemOutput{
						Id:     "POST",
						Output: make(chan OutputFrame),
					},
				},
			},
			expectedIds: []string{
				"GET", "POST",
			},
			expectFail: false,
		},
		{
			name: "run 3 different methods",
			rx: []*BatchItem{
				{
					Method:     GET,
					Iterations: 4,
					TargetURL:  svr.URL,
					Output: &ItemOutput{
						Id:     "GET",
						Output: make(chan OutputFrame),
					},
				},
				{
					Method:     POST,
					Iterations: 4,
					TargetURL:  svr.URL,
					Output: &ItemOutput{
						Id:     "POST",
						Output: make(chan OutputFrame),
					},
				},
				{
					Method:     PUT,
					Iterations: 4,
					TargetURL:  svr.URL,
					Output: &ItemOutput{
						Id:     "PUT",
						Output: make(chan OutputFrame),
					},
				},
			},
			expectedIds: []string{
				"GET", "POST", "PUT",
			},
			expectFail: false,
		},
		{
			name: "[expected fail] mismach an expected id ",
			rx: []*BatchItem{
				{
					Method:     GET,
					Iterations: 4,
					TargetURL:  svr.URL,
					Output: &ItemOutput{
						Id:     "GET",
						Output: make(chan OutputFrame),
					},
				},
				{
					Method:     POST,
					Iterations: 4,
					TargetURL:  svr.URL,
					Output: &ItemOutput{
						Id:     "POST",
						Output: make(chan OutputFrame),
					},
				},
				{
					Method:     POST,
					Iterations: 4,
					TargetURL:  svr.URL,
					Output: &ItemOutput{
						Id:     "PUT",
						Output: make(chan OutputFrame),
					},
				},
			},
			expectedIds: []string{
				"GET", "POST", "PUT",
			},
			expectFail: true,
		},
		{
			name: "run 3 instances, one writing to global",
			rx: []*BatchItem{
				{
					Method:     GET,
					Iterations: 4,
					TargetURL:  svr.URL,
					Output: &ItemOutput{
						Id:     "GET",
						Output: make(chan OutputFrame),
					},
				},
				{
					Method:     POST,
					Iterations: 4,
					TargetURL:  svr.URL,
					Output: &ItemOutput{
						Id:     "POST",
						Output: make(chan OutputFrame),
					},
				},
				{
					Method:     PUT,
					Iterations: 4,
					TargetURL:  svr.URL,
				},
			},
			expectedIds: []string{
				"GET", "POST", "PUT",
			},
			expectFail: false,
		},
		{
			name: "run 4 instances, 3 writing to global",
			rx: []*BatchItem{
				{
					Method:     GET,
					Iterations: 4,
					TargetURL:  svr.URL,
					Output: &ItemOutput{
						Id:     "GET",
						Output: make(chan OutputFrame),
					},
				},
				{
					Method:     POST,
					Iterations: 4,
					TargetURL:  svr.URL,
					Output: &ItemOutput{
						Id:     "POST",
						Output: make(chan OutputFrame),
					},
				},
				{
					Method:     PUT,
					Iterations: 4,
					TargetURL:  svr.URL,
				},
				{
					Method:     PATCH,
					Iterations: 4,
					TargetURL:  svr.URL,
				},
			},
			expectedIds: []string{
				"GET", "POST", "PUT", "PATCH",
			},
			expectFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := NewBatchRequest(tt.rx, false)
			go func() {
				for ret := range br.Output {
					t.Logf("Global Output: %v\n", string(ret.Data))
				}
			}()
			for ri, rx := range tt.rx {
				go func(ind int, i *BatchItem) {
					if i.Output == nil {
						if !tt.expectFail {
							return
						}
						t.Fail()
					}
					for ret := range i.Output.Output {
						if string(ret.Data) != tt.expectedIds[ind] && !tt.expectFail {
							t.Fail()
						}
					}
				}(ri, rx)
			}

			br.RunBatch()

		})
	}
}
