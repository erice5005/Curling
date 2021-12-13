package main

import (
	"log"
	"time"

	curling "github.com/erice5005/Curling"
)

func main() {
	br := curling.NewBatchRequest([]*curling.BatchItem{
		{
			TargetURL:  "https://httpbin.org/get",
			Iterations: 5,
			Method:     curling.GET,
			Delay:      250 * time.Millisecond,
		},
		{
			TargetURL:  "https://httpbin.org/post",
			Iterations: 5,
			Method:     curling.POST,
			Delay:      250 * time.Millisecond,
			Data:       "test",
		},
	}, false)

	go func() {
		for ret := range br.Output {
			log.Printf("RET: %v\n", string(ret.Data))
		}
	}()
	br.RunBatch()

}
