package core

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
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
