package core

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewUpstream(t *testing.T) {

	upstream1 := newUpstream(context.Background(), "http://test1.com", "http://test2.com")
	assert.IsType(t, &HttpUpstream{}, upstream1)

	upstream2 := newUpstream(context.Background(), "ws://test1.com", "ws://test2.com")
	assert.IsType(t, &WsUpstream{}, upstream2)

	assert.Panics(t, func() { newUpstream(context.Background(), "xxx://test1.com", "xxx://test2.com") })
}

func TestNewHttpUpstream(t *testing.T) {
	url1, err := url.Parse("http://test1.com")

	if err != nil {
		panic(err)
	}

	url2, err := url.Parse("http://test2.com")

	if err != nil {
		panic(err)
	}

	upstream1 := newHttpUpstream(context.Background(), url1, url2)
	assert.Equal(t, upstream1.url, "http://test1.com")
	assert.Equal(t, upstream1.oldTrieUrl, "http://test2.com")
}

func TestHttpHandle(t *testing.T) {
	url1, err := url.Parse("http://test1.com")

	if err != nil {
		panic(err)
	}

	url2, err := url.Parse("http://test2.com")

	if err != nil {
		panic(err)
	}

	upstream1 := newHttpUpstream(context.Background(), url1, url2)

	reqBodyBytes1 := []byte(fmt.Sprintf(`{"params": [], "method": "eth_blockNumber", "id": %d, "jsonrpc": "2.0"}`, time.Now().Unix()))
	req1, err := newRequest(reqBodyBytes1)

	if err != nil {
		panic(err)
	}

	bts, err := upstream1.handle(req1)

	if err != nil {
		panic(err)
	}

	assert.IsType(t, []byte{}, bts)
}

func TestNewWsUpstream(t *testing.T) {
	url1, err := url.Parse("http://test1.com")

	if err != nil {
		panic(err)
	}

	upstream1 := newWsStream(context.Background(), url1)
	assert.Equal(t, upstream1.url, "http://test1.com")
}

func TestWsHandle(t *testing.T) {
	url1, err := url.Parse("http://test1.com")

	if err != nil {
		panic(err)
	}

	upstream1 := newWsStream(context.Background(), url1)

	reqBodyBytes1 := []byte(fmt.Sprintf(`{"params": [], "method": "eth_blockNumber", "id": %d, "jsonrpc": "2.0"}`, time.Now().Unix()))
	req1, err := newRequest(reqBodyBytes1)

	if err != nil {
		panic(err)
	}

	bts, err := upstream1.handle(req1)

	assert.NotEqual(t, nil, err)

	url2, err := url.Parse("ws://test1.com")

	if err != nil {
		panic(err)
	}

	upstream2 := newWsStream(context.Background(), url2)

	reqBodyBytes2 := []byte(fmt.Sprintf(`{"params": [], "method": "eth_blockNumber", "id": %d, "jsonrpc": "2.0"}`, time.Now().Unix()))
	req2, err := newRequest(reqBodyBytes2)

	if err != nil {
		panic(err)
	}

	bts, err = upstream2.handle(req2)

	assert.Equal(t, TimeoutError, err)

	assert.IsType(t, []byte{}, bts)
}

func TestWsRun(t *testing.T) {
	url2, err := url.Parse("ws://test1.com")

	if err != nil {
		panic(err)
	}

	upstream2 := newWsStream(context.Background(), url2)

	timeout := time.After(5 * time.Second)
	done := make(chan bool)
	go func() {
		upstream2.run(context.Background())
		done <- true
	}()

	select {
	case <-timeout:
		assert.True(t, true)
	case <-done:
		assert.True(t, true)
	}
}
