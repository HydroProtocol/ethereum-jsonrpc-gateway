package core

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/HydroProtocol/ethereum-jsonrpc-gateway/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetBlockNumberRequest(t *testing.T) {
	req := getBlockNumberRequest()
	assert.Equal(t, "eth_blockNumber", req.data.Method)
}

func TestIsOldTrieRequest(t *testing.T) {
	logger := log.New(os.Stdout, fmt.Sprintf("[id: %v] ", utils.RandStringRunes(8)), log.LstdFlags)

	reqBodyBytes1 := []byte(fmt.Sprintf(`{"params": [], "method": "eth_blockNumber", "id": %d, "jsonrpc": "2.0"}`, time.Now().Unix()))

	var data1 RequestData
	_ = json.Unmarshal(reqBodyBytes1, &data1)

	req1 := &Request{
		logger:   logger,
		data:     &data1,
		reqBytes: reqBodyBytes1,
	}

	assert.Equal(t, false, req1.isOldTrieRequest(0))
	assert.Equal(t, false, req1.isOldTrieRequest(1))

	reqBodyBytes2 := []byte(fmt.Sprintf(`{"params": [], "method": "eth_call", "id": %d, "jsonrpc": "2.0"}`, time.Now().Unix()))

	var data2 RequestData
	_ = json.Unmarshal(reqBodyBytes2, &data2)

	req2 := &Request{
		logger:   logger,
		data:     &data2,
		reqBytes: reqBodyBytes2,
	}

	assert.Equal(t, false, req2.isOldTrieRequest(1))

	reqBodyBytes3 := []byte(fmt.Sprintf(`{"params": ["testParams0", "1"], "method": "eth_call", "id": %d, "jsonrpc": "2.0"}`, time.Now().Unix()))

	var data3 RequestData
	_ = json.Unmarshal(reqBodyBytes3, &data3)

	req3 := &Request{
		logger:   logger,
		data:     &data3,
		reqBytes: reqBodyBytes3,
	}

	assert.Equal(t, true, req3.isOldTrieRequest(10000))
	reqBodyBytes4 := []byte(fmt.Sprintf(`{"params": [1, 1], "method": "eth_call", "id": %d, "jsonrpc": "2.0"}`, time.Now().Unix()))

	var data4 RequestData
	_ = json.Unmarshal(reqBodyBytes4, &data4)

	req4 := &Request{
		logger:   logger,
		data:     &data4,
		reqBytes: reqBodyBytes4,
	}

	assert.Equal(t, true, req4.isOldTrieRequest(10000))
}
