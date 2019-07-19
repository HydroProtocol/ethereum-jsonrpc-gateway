package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var TimeoutError = fmt.Errorf("timeout error")
var AllUpstreamsFailedError = fmt.Errorf("all upstream requests are failed")

type Request struct {
	logger *log.Logger
	body   []byte
	r      *http.Request
	w      http.ResponseWriter
	data   *RequestData
}

func newRequest(r *http.Request, w http.ResponseWriter) *Request {
	logger := log.New(os.Stdout, fmt.Sprintf("[id: %v] ", randStringRunes(8)), log.LstdFlags)
	reqBodyBytes, _ := ioutil.ReadAll(r.Body)

	var data RequestData
	_ = json.Unmarshal(reqBodyBytes, &data)

	logger.Printf("New, method: %s\n", data.Method)
	//logger.Printf("Request Body: %s\n", string(reqBodyBytes))
	return &Request{
		logger: logger,
		r:      r,
		w:      w,
		body:   reqBodyBytes,
		data:   &data,
	}
}

func (r *Request) returnError(reason interface{}) {
	r.w.WriteHeader(500)

	bts, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      r.data.ID,
		"error": map[string]interface{}{
			"code":    -32602,
			"message": reason,
		},
	})

	_, _ = r.w.Write(bts)
}

func (r *Request) valid() error {

	err := isValidCall(r.data)

	if err != nil {
		r.logger.Printf("not valid, skip\n")

		return err
	}

	return nil
}

func (r *Request) naiveProxy() error {
	upstream := upstreams[0]

	upstreamReq, _ := http.NewRequest("POST", upstream, bytes.NewReader(r.body))
	upstreamReq.Header.Set("Content-Type", "application/json")

	res, err := httpClient.Do(upstreamReq)

	if err != nil {
		return err
	}

	pipeRes(r.w, res)

	return nil
}

func (r *Request) raceProxy() error {
	startAt := time.Now()

	defer func() {
		//utils.Monitor.MonitorTime("geth_gateway", float64(time.Since(startAt)) / 1000000)
	}()

	successfulResponse := make(chan *http.Response, len(upstreams))
	failedResponse := make(chan *http.Response, len(upstreams))
	errorResponseUpstreams := make(chan string, len(upstreams))

	for _, upstream := range upstreams {
		go func(upstream string) {
			defer func() {
				if err := recover(); err != nil {
					r.logger.Printf("%v Upstream %s failed, err: %v\n", time.Now().Sub(startAt), upstream, err)
					errorResponseUpstreams <- upstream
				}
			}()

			upstreamReq, _ := http.NewRequest("POST", upstream, bytes.NewReader(r.body))
			upstreamReq.Header.Set("Content-Type", "application/json")

			res, err := httpClient.Do(upstreamReq)

			if err != nil {
				r.logger.Printf("%vms Upstream: %v, Error: %v\n", time.Now().Sub(startAt), upstream, err)
				failedResponse <- nil
				return
			}

			resBody := strings.TrimSpace(string(peekResponseBody(res)))

			diff := time.Now().Sub(startAt)
			if res.StatusCode >= 200 && res.StatusCode < 300 && noErrorFieldInJSON(resBody) {
				r.logger.Printf("%v Upstream: %v Success[%d], Body: %v\n", diff, upstream, res.StatusCode, resBody)
				successfulResponse <- res
			} else {
				r.logger.Printf("%v Upstream: %v Failed[%d], Body: %v\n", diff, upstream, res.StatusCode, resBody)
				failedResponse <- res
			}
		}(upstream)
	}

	failedCount := 0
	errorCount := 0
	var failedResponses = make([]*http.Response, 0, len(upstreams))

	for failedCount+errorCount < len(upstreams) {
		select {
		case <-time.After(time.Second * 10):
			timeout(r.w)
			r.logger.Printf("%v Final Timeout\n", time.Now().Sub(startAt))
			return TimeoutError
		case res := <-successfulResponse:
			pipeRes(r.w, res)
			r.logger.Printf("%v Final Success\n", time.Now().Sub(startAt))
			return nil
		case res := <-failedResponse:
			failedResponses = append(failedResponses, res)
			failedCount++
		case <-errorResponseUpstreams:
			errorCount++
		}
	}

	pipeRes(r.w, failedResponses[0])
	r.logger.Printf("%v Final Failed\n", time.Now().Sub(startAt))

	//utils.Monitor.MonitorCount("geth_gateway_fail")
	return AllUpstreamsFailedError
}
