package main

import (
	_ "github.com/joho/godotenv/autoload"
)

import (
	"bytes"
	"github.com/TV4/graceful"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

var httpClient *http.Client

const (
	maxIdleConnections int = 100
	requestTimeout     int = 5
)

func init() {
	httpClient = createHTTPClient()
	rand.Seed(time.Now().UnixNano())
}

// createHTTPClient for connection re-use
func createHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: maxIdleConnections,
		},
		Timeout: time.Duration(requestTimeout) * time.Second,
	}
}

type Server struct{}

var upstreams = strings.Split(os.Getenv("UPSTREAMS"), ",")

func pipeRes(rw http.ResponseWriter, res *http.Response) {
	rw.WriteHeader(res.StatusCode)
	rw.Header().Set("Content-Type", res.Header.Get("Content-Type"))
	rw.Header().Set("Content-Length", res.Header.Get("Content-Length"))
	_, _ = io.Copy(rw, res.Body)
	_ = res.Body.Close()
}

func timeout(w http.ResponseWriter) {
	w.WriteHeader(504)
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
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if req.Method != http.MethodPost {
		w.WriteHeader(400)
		_, _ = w.Write([]byte("Method Should Be POST"))
		return
	}

	proxyRequest := newRequest(req, w)
	err := proxyRequest.valid()

	if err != nil {
		proxyRequest.returnError(err.Error())
		return
	}

	upstreamsCount := len(upstreams)

	if upstreamsCount == 1 {
		err = proxyRequest.naiveProxy()
	} else {
		err = proxyRequest.raceProxy()
	}

	if err != nil {
		proxyRequest.returnError(err.Error())
	}

}

func main() {
	if len(upstreams) == 0 {
		panic("need upstreams")
	}

	hs := &http.Server{Addr: ":3005", Handler: &Server{}}

	go graceful.Shutdown(hs)

	log.Printf("Listening on http://0.0.0.0%s\n", hs.Addr)

	//router := httprouter.New()
	//router.Handler("GET", "/metrics", utils.Monitor.Handler())
	//go http.ListenAndServe(fmt.Sprintf(":%s", "9091"), router)

	if err := hs.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
