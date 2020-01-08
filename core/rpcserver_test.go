package core

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateHTTPClient(t *testing.T) {
	assert.IsType(t, &http.Client{}, createHTTPClient())
}

func TestGetErrorResponseBytes(t *testing.T) {
	bts := getErrorResponseBytes(1, "test msg")

	assert.IsType(t, "{\"error\":{\"code\":-32602,\"message\":\"test msg\"},\"id\":1,\"jsonrpc\":\"2.0\"}", string(bts))
}

func TestServerWS(t *testing.T) {
	assert.IsType(t, &http.Client{}, createHTTPClient())
}
