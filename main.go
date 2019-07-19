package main

import _ "github.com/joho/godotenv/autoload"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/TV4/graceful"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
	"github.com/julienschmidt/httprouter"
)

var httpClient *http.Client

const (
	maxIdleConnections int = 20
	requestTimeout     int = 5
)

func init() {
	httpClient = createHTTPClient()
	rand.Seed(time.Now().UnixNano())
}

// createHTTPClient for connection re-use
func createHTTPClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: maxIdleConnections,
		},
		Timeout: time.Duration(requestTimeout) * time.Second,
	}

	return client
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func noErrorFieldInJSON(jsonStr string) bool {
	var tmp map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &tmp)

	if err != nil {
		log.Printf("decode json string failed, %v, %v\n", jsonStr, err)
		return false
	}

	if tmp["error"] == nil {
		return true
	}

	return false
}

// Server ..
type Server struct{}

var upstreams = strings.Split(os.Getenv("UPSTREAMS"), ",")

func pipeRes(rw http.ResponseWriter, res *http.Response) {
	rw.WriteHeader(res.StatusCode)
	rw.Header().Set("Content-Type", res.Header.Get("Content-Type"))
	rw.Header().Set("Content-Length", res.Header.Get("Content-Length"))
	io.Copy(rw, res.Body)
	res.Body.Close()
}

func timeout(w http.ResponseWriter) {
	w.WriteHeader(504)
}

func fail(w http.ResponseWriter) {
	w.WriteHeader(500)
	w.Write([]byte("All upstream requests are failed"))
}

func peekResponseBody(res *http.Response) []byte {
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	res.Body = ioutil.NopCloser(bytes.NewReader(buf))
	return buf
}

func (h *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.WriteHeader(400)
		w.Write([]byte("Method Should Be POST"))
		return
	}

	startAt := time.Now()
	defer func(){
		//utils.Monitor.MonitorTime("geth_gateway", float64(time.Since(startAt)) / 1000000)
	}()

	u1 := randStringRunes(8)
	reqBodyBytes, _ := ioutil.ReadAll(req.Body)
	logger := log.New(os.Stdout, fmt.Sprintf("[id: %v] ", u1), log.LstdFlags)

	logger.Printf("New Request Received\n")
	logger.Printf("Request Body: %s\n", string(reqBodyBytes))

	successfulResponse := make(chan *http.Response, len(upstreams))
	failedResponse := make(chan *http.Response, len(upstreams))
	errorResponseUpstreams := make(chan string, len(upstreams))

	for _, upstream := range upstreams {
		go func(upstream string) {
			defer func() {
				if err := recover(); err != nil {
					logger.Printf("%v Upstream %s failed, err: %v\n", time.Now().Sub(startAt), upstream, err)
					errorResponseUpstreams <- upstream
				}
			}()

			upstreamReq, _ := http.NewRequest("POST", upstream, bytes.NewReader(reqBodyBytes))
			upstreamReq.Header.Set("Content-Type", "application/json")

			res, err := httpClient.Do(upstreamReq)

			if err != nil {
				logger.Printf("%vms Upstream: %v, Error: %v\n", time.Now().Sub(startAt), upstream, err)
				failedResponse <- nil
				return
			}

			resBody := strings.TrimSpace(string(peekResponseBody(res)))

			diff := time.Now().Sub(startAt)
			if res.StatusCode >= 200 && res.StatusCode < 300 && noErrorFieldInJSON(resBody) {
				logger.Printf("%v Upstream: %v Success[%d], Body: %v\n", diff, upstream, res.StatusCode, resBody)
				successfulResponse <- res
			} else {
				logger.Printf("%v Upstream: %v Failed[%d], Body: %v\n", diff, upstream, res.StatusCode, resBody)
				failedResponse <- res
			}
		}(upstream)
	}

	failedCount := 0
	errorCount := 0
	failedResponses := []*http.Response{}

	for failedCount+errorCount < len(upstreams) {
		select {
		case <-time.After(time.Second * 10):
			timeout(w)
			logger.Printf("%v Final Timeout\n", time.Now().Sub(startAt))
			return
		case res := <-successfulResponse:
			pipeRes(w, res)
			logger.Printf("%v Final Success\n", time.Now().Sub(startAt))
			return
		case res := <-failedResponse:
			failedResponses = append(failedResponses, res)
			failedCount++
		case <-errorResponseUpstreams:
			errorCount++
		}
	}

	if len(failedResponses) > 0 {
		pipeRes(w, failedResponses[0])
	} else {
		fail(w)
	}

	//utils.Monitor.MonitorCount("geth_gateway_fail")
	logger.Printf("%v Final Failed\n", time.Now().Sub(startAt))
}

func main() {
	hs := &http.Server{Addr: ":3000", Handler: &Server{}}

	go graceful.Shutdown(hs)

	log.Printf("Listening on http://0.0.0.0%s\n", hs.Addr)

	router := httprouter.New()
	//router.Handler("GET", "/metrics", utils.Monitor.Handler())
	go http.ListenAndServe(fmt.Sprintf(":%s", "9091"), router)

	if err := hs.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
