package core

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCreateHTTPClient(t *testing.T) {
	assert.IsType(t, &http.Client{}, createHTTPClient())
}

func TestGetErrorResponseBytes(t *testing.T) {
	bts := getErrorResponseBytes(1, "test msg")

	assert.Equal(t, "{\"error\":{\"code\":-32602,\"message\":\"test msg\"},\"id\":1,\"jsonrpc\":\"2.0\"}", string(bts))
}

func TestServeHTTP(t *testing.T) {
	httpServer := &http.Server{Addr: ":3005", Handler: &Server{}}

	go func() {
		time.Sleep(5 * time.Second)

		if err := httpServer.Shutdown(context.Background()); err != nil {
			logrus.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
	}()

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		logrus.Fatal(err)
	}

	assert.Equal(t, 1, 1)
}
