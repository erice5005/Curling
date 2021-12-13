package curling

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type BatchRequest struct {
	Items  []*BatchItem
	Output chan OutputFrame
}

func NewBatchRequest(items []*BatchItem, initOutput bool) BatchRequest {
	br := BatchRequest{
		Items: items,
	}
	br.Output = make(chan OutputFrame, br.GetTotalIterations()) //Buffer output to accomodate all the iterations
	return br
}

func (br *BatchRequest) SetOutput(c chan OutputFrame) {
	br.Output = c
}

func (br *BatchRequest) GetTotalIterations() int {
	out := 0
	for _, i := range br.Items {
		out += i.Iterations
	}

	return out
}

func (br *BatchRequest) Add(bi *BatchItem) {
	if i := itemInSlice(bi, br.Items); i == -1 { // Unique items only
		br.Items = append(br.Items, bi)
	}
}

func itemInSlice(bi *BatchItem, haystack []*BatchItem) int {
	for ki, kx := range haystack {
		if kx.Id == bi.Id {
			return ki
		}
	}
	return -1
}

func (br *BatchRequest) RunBatch() {
	var wg sync.WaitGroup
	for _, i := range br.Items {
		wg.Add(1)
		go func(wg *sync.WaitGroup, bi *BatchItem, o chan OutputFrame) {
			defer wg.Done()
			bi.Exec(o)
		}(&wg, i, br.Output)
	}
	wg.Wait()
	close(br.Output)
}

type ItemOutput struct {
	Id     string
	Output chan OutputFrame
}

type OutputFrame struct {
	Data    []byte
	Headers map[string][]string
	Err     error
}

type BatchItem struct {
	Id             uuid.UUID
	TargetURL      string
	Headers        map[string]string
	Data           interface{}
	Method         Method
	Iterations     int
	Delay          time.Duration
	IterationsDone int
	Output         *ItemOutput
}

func NewBatchItem(Method Method, TargetURL string, Headers map[string]string) *BatchItem {
	bi := &BatchItem{
		Id:         uuid.New(),
		TargetURL:  TargetURL,
		Method:     Method,
		Headers:    Headers,
		Iterations: 1,
	}
	return bi
}

func (bi *BatchItem) SetIterations(iter int) {
	bi.Iterations = iter
}

func (bi *BatchItem) SetOutput(it *ItemOutput) {
	bi.Output = it
}

func (bi *BatchItem) GetOutput() ItemOutput {
	return *bi.Output
}

func (bi *BatchItem) Exec(globalOutput chan<- OutputFrame) {
	if bi.Output != nil {
		globalOutput = bi.Output.Output
	}
	for bi.IterationsDone = 0; bi.IterationsDone < bi.Iterations; bi.IterationsDone++ {
		outputFrame := OutputFrame{}
		outputFrame.Data, outputFrame.Headers, outputFrame.Err = New(bi.Method, bi.TargetURL, bi.Headers).Do(bi.Data)
		globalOutput <- outputFrame
		time.Sleep(bi.Delay)
	}
	if bi.Output != nil {
		close(bi.Output.Output)
	}
}
