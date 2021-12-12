package main

import (
	"log"

	curling "github.com/erice5005/Curling"
)

func main() {
	body, headers, err := curling.New(curling.GET, "https://httpbin.org/get", nil).Do(nil)
	if err != nil {
		panic(err)
	}
	log.Printf("Headers: %v\n", headers)
	log.Printf("Body: %v\n", string(body))
}
