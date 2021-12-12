package curling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

type Method int64

const (
	GET   Method = 0
	POST  Method = 1
	PUT   Method = 2
	PATCH Method = 3
)

func (m Method) String() string {
	switch m {
	case GET:
		return "GET"
	case POST:
		return "POST"
	case PUT:
		return "PUT"
	case PATCH:
		return "PATCH"
	}

	return ""
}

type Request struct {
	Method    Method
	TargetURL string
	Headers   map[string]string
}

func New(method Method, targetURL string, headers map[string]string) Request {
	r := Request{
		Method:    method,
		TargetURL: targetURL,
		Headers:   headers,
	}

	return r
}

// Do ... main execution of the package. Can accept either JSON or
func (r Request) Do(dataset interface{}) ([]byte, map[string][]string, error) {
	var postbody io.Reader
	var err error
	if r.Method == POST {
		postbody, err = getReaderForType(dataset)
		if err != nil {
			return nil, nil, fmt.Errorf("error setting postbody: %w", err)
		}
	}

	req, err := http.NewRequest(r.Method.String(), r.TargetURL, postbody)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %w", err)
	}
	for key, val := range r.Headers {
		req.Header.Add(key, val)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error executing requesst: %w", err)
	}

	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	headers := make(map[string][]string)
	for name, value := range res.Header {
		headers[name] = value
	}

	return body, headers, nil
}

func getReaderForType(data interface{}) (io.Reader, error) {
	if data == nil {
		return nil, nil
	}

	switch reflect.TypeOf(data).Kind() {
	case reflect.String:
		return strings.NewReader(data.(string)), nil
	case reflect.Struct:
		marshed, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(marshed), nil
	}

	return nil, nil
}
