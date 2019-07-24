package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"
	"os/signal"
	"strings"
	"syscall"
)

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
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

var upstreams []Upstream
var strategy IStrategy
var methodLimitation = os.Getenv("LIMITATION") == "true"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // kong should take care of cors
	},
}

func (h *Server) ServerWS(conn *websocket.Conn) error {
	defer conn.Close()

	for {
		messageType, r, err := conn.NextReader()
		if err != nil {
			return err
		}

		w, err := conn.NextWriter(messageType)

		if err != nil {
			return err
		}

		reqBodyBytes, _ := ioutil.ReadAll(r)
		proxyRequest, err := newRequest(reqBodyBytes)

		if err != nil {
			return err
		}

		bts, err := strategy.handle(proxyRequest)

		if err != nil {
			bts = getErrorResponseBytes(proxyRequest.data.ID, err.Error())
		}

		if _, err := w.Write(bts); err != nil {
			return err
		}

		if err := w.Close(); err != nil {
			return err
		}
	}
}

func getErrorResponseBytes(id interface{}, reason interface{}) []byte {
	bts, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    -32602,
			"message": reason,
		},
	})

	return bts
}

func (h *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/ws" {
		conn, err := upgrader.Upgrade(w, req, nil)

		if err != nil {
			logrus.Error(err)
			return
		}

		_ = h.ServerWS(conn)
		return
	}

	w.Header().Set("Content-Type", "application/json")
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

	reqBodyBytes, _ := ioutil.ReadAll(req.Body)
	proxyRequest, err := newRequest(reqBodyBytes)

	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write(getErrorResponseBytes(proxyRequest.data.ID, err.Error()))
		return
	}

	bts, err := strategy.handle(proxyRequest)

	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write(getErrorResponseBytes(proxyRequest.data.ID, err.Error()))
		return
	}

	_, _ = w.Write(bts)
}

func initialize(ctx context.Context) {
	for _, url := range strings.Split(os.Getenv("UPSTREAMS"), ",") {
		upstreams = append(upstreams, newUpstream(ctx, url))
	}

	if len(upstreams) == 0 {
		panic("need upstreams")
	}

	s := os.Getenv("STRATEGY")
	switch s {
	case "NAIVE":
		if len(upstreams) > 1 {
			panic(fmt.Errorf("naive proxy strategy require exact 1 upstream"))
		}
		strategy = newNaiveProxy()
	case "RACE":
		if len(upstreams) < 2 {
			panic(fmt.Errorf("race proxy strategy require more than 1 upstream"))
		}
		strategy = newRaceProxy()
	case "FALLBACK":
		if len(upstreams) < 2 {
			panic(fmt.Errorf("fallback proxy strategy require more than 1 upstream"))
		}
		strategy = newFallbackProxy()
	default:
		panic(fmt.Errorf("blank of unsupported strategy: %s", s))
	}

	initLimitation()
}

func setLogLevel() {
	var level logrus.Level

	switch os.Getenv("LOG_LEVEL") {
	case "DEBUG":
		level = logrus.DebugLevel
	case "ERROR":
		level = logrus.ErrorLevel
	case "FATAL":
		level = logrus.FatalLevel
	case "TRACE":
		level = logrus.TraceLevel
	case "PANIC":
		level = logrus.PanicLevel
	default:
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
}

func waitExitSignal(ctxStop context.CancelFunc) {
	var exitSignal = make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGTERM)
	signal.Notify(exitSignal, syscall.SIGINT)

	<-exitSignal
	logrus.Info("Stopping...")
	ctxStop()
}

func run() int {
	ctx, stop := context.WithCancel(context.Background())
	go waitExitSignal(stop)

	setLogLevel()
	initialize(ctx)

	hs := &http.Server{Addr: ":3005", Handler: &Server{}}

	// http server graceful shutdown
	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := hs.Shutdown(shutdownCtx); err != nil {
			logrus.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
	}()

	logrus.Infof("Listening on http://0.0.0.0%s\n", hs.Addr)

	// TODO monitoring
	// router := httprouter.New()
	// router.Handler("GET", "/metrics", utils.Monitor.Handler())
	// go http.ListenAndServe(fmt.Sprintf(":%s", "9091"), router)

	if err := hs.ListenAndServe(); err != http.ErrServerClosed {
		logrus.Fatal(err)
	}

	logrus.Info("Stopped")
	return 0
}

func main() {
	os.Exit(run())
}
