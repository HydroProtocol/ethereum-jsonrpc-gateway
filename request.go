package main

import (
	"encoding/json"
	"fmt"
	"git.ddex.io/lib/monitor"
	"log"
	"os"
)

var TimeoutError = fmt.Errorf("timeout error")
var AllUpstreamsFailedError = fmt.Errorf("all upstream requests are failed")

type Request struct {
	logger   *log.Logger
	data     *RequestData
	reqBytes []byte
}

func newRequest(reqBodyBytes []byte) (*Request, error) {
	logger := log.New(os.Stdout, fmt.Sprintf("[id: %v] ", randStringRunes(8)), log.LstdFlags)

	var data RequestData
	_ = json.Unmarshal(reqBodyBytes, &data)

	logger.Printf("New, method: %s\n", data.Method)
	//logger.Printf("Request Body: %s\n", string(reqBodyBytes))

	req := &Request{
		logger:   logger,
		data:     &data,
		reqBytes: reqBodyBytes,
	}

	// method limit, for directly external access
	err := req.valid()

	if err != nil {
		return req, err
	}

	return req, nil
}

func (r *Request) valid() error {

	if !currentRunningConfig.MethodLimitationEnabled {
		return nil
	}

	err := isValidCall(r.data)

	if err != nil {
		r.logger.Printf("not valid, skip\n")
		monitor.Count("block_call", r.data.Method)
		return err
	}

	return nil
}
