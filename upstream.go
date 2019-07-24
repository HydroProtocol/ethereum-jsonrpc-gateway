package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// the handle function will execute concurrently
type Upstream interface {
	handle(*Request) ([]byte, error)
}

type wsProxyRequest struct {
	*Request
	id       int64
	resBytes chan []byte
}

type wsProxyResponse struct {
	ID int64 `json:"id"`
}

type WsUpstream struct {
	url          string
	requestQueue chan *wsProxyRequest
	nextID       int64     // proxy request id
	requests     *sync.Map // proxy request id => proxy request
}

type HttpUpstream struct {
	ctx context.Context
	url string
}

func newUpstream(ctx context.Context, urlString string) Upstream {
	u, err := url.Parse(urlString)

	if err != nil {
		panic(err)
	}

	if u.Scheme == "http" || u.Scheme == "https" {
		return newHttpUpstream(ctx, u)
	} else if u.Scheme == "ws" || u.Scheme == "wss" {
		return newWsStream(ctx, u)
	} else {
		panic(fmt.Errorf("unsuportted url schema %s", u.Scheme))
	}
}

func (u *HttpUpstream) handle(request *Request) ([]byte, error) {
	upstreamReq, _ := http.NewRequest("POST", u.url, bytes.NewReader(request.reqBytes))
	upstreamReq.Header.Set("Content-Type", "application/json")

	res, err := httpClient.Do(upstreamReq)

	if err != nil {
		return nil, err
	}

	bts, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	return bts, nil
}

func (u *WsUpstream) handle(request *Request) ([]byte, error) {
	proxyRequest := &wsProxyRequest{
		request,
		atomic.AddInt64(&u.nextID, 1),
		make(chan []byte),
	}

	u.requests.Store(proxyRequest.id, proxyRequest)
	defer u.requests.Delete(proxyRequest.id)

	select {
	case u.requestQueue <- proxyRequest:
	case <-time.After(5 * time.Second): // TODO use a configurable timeout
		return nil, TimeoutError
	}

	select {
	case res := <-proxyRequest.resBytes:
		return res, nil
	case <-time.After(5 * time.Second): // TODO use a configurable timeout
		return nil, TimeoutError
	}
}

func (u *WsUpstream) run(ctx context.Context) {
	logrus.Debugf("ws %s run", u.url)
	defer logrus.Debugf("ws %s run exit", u.url)

	for {
		conn, _, err := websocket.DefaultDialer.Dial(u.url, nil)

		if err != nil {
			seconds := 5 // TODO configurable
			logrus.Errorf("ws upstream %s %v, will retry after %d seconds", u.url, err, seconds)

			select {
			case <-ctx.Done():
				// global stop
				return
			case <-time.After(time.Second * time.Duration(seconds)):
				continue
			}

		}

		logrus.Infof("ws upstream %s connected", u.url)
		u.runConn(ctx, conn)

		select {
		case <-ctx.Done():
			// global stop
			return
		}
	}
}

// return the connection context
func (u *WsUpstream) runConn(ctx context.Context, conn *websocket.Conn) {
	defer conn.Close()

	// connContext is for current connection
	// any error occurs, the context will be cancelled
	connContext, done := context.WithCancel(ctx)

	// request loop
	go func() {
		logrus.Debugf("conn request loop start")
		defer logrus.Debugf("conn request loop stop")
		defer done()
		for {
			select {
			case <-connContext.Done():
				// if the conn is invalid, exit
				return
			case wsProxyRequest := <-u.requestQueue:
				// use proxy ID
				wsProxyRequest.Request.data.ID = wsProxyRequest.id

				bts, _ := json.Marshal(wsProxyRequest.Request.data)
				err := conn.WriteMessage(websocket.TextMessage, bts)

				if err != nil {
					logrus.Errorf("write request to upstream failed %v", err)
					return
				}
			}

		}
	}()

	// response loop
	go func() {
		logrus.Debugf("conn response loop start")
		defer logrus.Debugf("conn response loop stop")
		defer done()

		for {
			t, p, err := conn.ReadMessage()

			if err != nil {
				logrus.Errorf("read response from upstream failed %v", err)
				break
			}

			if t != websocket.TextMessage {
				logrus.Infof("not a text message %v", p)
				continue
			}

			var res wsProxyResponse
			_ = json.Unmarshal(p, &res)

			if r, exist := u.requests.Load(res.ID); exist {
				if req, ok := r.(*wsProxyRequest); ok {
					req.resBytes <- p
				}
			}
		}
	}()

	<-connContext.Done()
}

func newHttpUpstream(ctx context.Context, url *url.URL) *HttpUpstream {
	return &HttpUpstream{
		ctx: ctx,
		url: url.String(),
	}
}

func newWsStream(ctx context.Context, url *url.URL) *WsUpstream {
	upstream := &WsUpstream{
		url:          url.String(),
		requestQueue: make(chan *wsProxyRequest),
		nextID:       time.Now().Unix(),
		requests:     &sync.Map{},
	}

	logrus.Infof("new upstream %s", url)
	go upstream.run(ctx)

	return upstream
}
