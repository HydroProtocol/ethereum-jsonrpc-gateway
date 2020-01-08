package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var httpClient *http.Client

const (
	maxIdleConnections int = 200
	requestTimeout     int = 10
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

type Config struct {
	Upstreams               []string `json:"upstreams"`
	Strategy                string   `json:"strategy"`
	MethodLimitationEnabled bool     `json:"methodLimitationEnabled"`
	AllowedMethods          []string `json:"allowedMethods"`
	ContractWhitelist       []string `json:"contractWhitelist"`
}

type RunningConfig struct {
	ctx                     context.Context
	stop                    context.CancelFunc
	Upstreams               []Upstream
	Strategy                IStrategy
	MethodLimitationEnabled bool
	allowedMethods          map[string]bool
	allowedCallContracts    map[string]bool
}

var currentRunningConfig *RunningConfig

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

		bts, err := currentRunningConfig.Strategy.handle(proxyRequest)

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
	w.Header().Set("Access-Control-Allow-Origin", "*")

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
		logrus.Errorf("Req from %s %s 500 %s", req.RemoteAddr, proxyRequest.data.Method, err.Error())
		return
	}

	bts, err := currentRunningConfig.Strategy.handle(proxyRequest)

	var isArchiveRequestText string
	if proxyRequest.isArchiveDataRequest {
		isArchiveRequestText = "(ArchiveData)"
		logrus.Info(string(proxyRequest.reqBytes))
	}

	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write(getErrorResponseBytes(proxyRequest.data.ID, err.Error()))
		logrus.Errorf("Req%s from %s %s 500 %s", isArchiveRequestText, req.RemoteAddr, proxyRequest.data.Method, err.Error())
		return
	}

	_, _ = w.Write(bts)
	logrus.Infof("Req%s from %s %s 200", isArchiveRequestText, req.RemoteAddr, proxyRequest.data.Method)
}

func buildRunningConfigFromConfig(parentContext context.Context, cfg *Config) (*RunningConfig, error) {
	ctx, stop := context.WithCancel(parentContext)

	rcfg := &RunningConfig{
		ctx:  ctx,
		stop: stop,
	}

	for _, url := range cfg.Upstreams {

		// hack, refactor this sometime

		var primaryUrl string
		var oldTrieUrl string

		if strings.Contains(url, ",") {
			urls := strings.Split(url, ",")
			primaryUrl = urls[0]
			oldTrieUrl = urls[1]
		} else {
			primaryUrl = url
			oldTrieUrl = url
		}

		rcfg.Upstreams = append(rcfg.Upstreams, newUpstream(ctx, primaryUrl, oldTrieUrl))
	}

	if len(rcfg.Upstreams) == 0 {
		return nil, fmt.Errorf("need upstreams")
	}

	switch cfg.Strategy {
	case "NAIVE":
		if len(rcfg.Upstreams) > 1 {
			panic(fmt.Errorf("naive proxy strategy require exact 1 upstream"))
		}
		rcfg.Strategy = newNaiveProxy()
	case "RACE":
		if len(rcfg.Upstreams) < 2 {
			panic(fmt.Errorf("race proxy strategy require more than 1 upstream"))
		}
		rcfg.Strategy = newRaceProxy()
	case "FALLBACK":
		if len(rcfg.Upstreams) < 2 {
			panic(fmt.Errorf("fallback proxy strategy require more than 1 upstream"))
		}
		rcfg.Strategy = newFallbackProxy()
	default:
		return nil, fmt.Errorf("blank of unsupported strategy: %s", cfg.Strategy)
	}

	rcfg.MethodLimitationEnabled = cfg.MethodLimitationEnabled

	rcfg.allowedMethods = make(map[string]bool)
	for i := 0; i < len(cfg.AllowedMethods); i++ {
		rcfg.allowedMethods[cfg.AllowedMethods[i]] = true
	}

	rcfg.allowedCallContracts = make(map[string]bool)
	for i := 0; i < len(cfg.ContractWhitelist); i++ {
		rcfg.allowedCallContracts[strings.ToLower(cfg.ContractWhitelist[i])] = true
	}

	return rcfg, nil
}

func waitExitSignal(ctxStop context.CancelFunc) {
	var exitSignal = make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGTERM)
	signal.Notify(exitSignal, syscall.SIGINT)

	<-exitSignal

	logrus.Info("Stopping...")
	ctxStop()
}

func Run() int {

	ctx, stop := context.WithCancel(context.Background())
	go waitExitSignal(stop)

	config := &Config{}

	logrus.Info("load config from file")
	bts, err := ioutil.ReadFile("./config.json")

	if err != nil {
		logrus.Fatal(err)
	}

	_ = json.Unmarshal(bts, config)

	currentRunningConfig, err = buildRunningConfigFromConfig(ctx, config)

	// test reload config
	//go func() {
	//	time.Sleep(5 * time.Second)
	//
	//	oldRunningConfig := currentRunningConfig
	//	newRcfg, err := buildRunningConfigFromConfig(ctx, config)
	//
	//	if err == nil {
	//		currentRunningConfig = newRcfg
	//		oldRunningConfig.stop()
	//		logrus.Info("running config changes successfully")
	//	} else {
	//		logrus.Info("running config changes failed, err: %+v", err)
	//	}
	//}()

	if err != nil {
		logrus.Fatal(err)
	}

	httpServer := &http.Server{Addr: ":3005", Handler: &Server{}}

	// http server graceful shutdown
	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logrus.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
	}()

	logrus.Infof("Listening on http://0.0.0.0%s\n", httpServer.Addr)

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		logrus.Fatal(err)
	}

	logrus.Info("Stopped")
	return 0
}
