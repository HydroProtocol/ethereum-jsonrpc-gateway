package core

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIsAllowedMethod(t *testing.T) {
	var testConfigStr1 = `{
		"_upstreams": "support http, https, ws, wss",
		"upstreams": [
		  "https://ropsten.infura.io/v3/83438c4dcf834ceb8944162688749707"
		],
	  
		"_strategy": "support NAIVE, RACE, FALLBACK",
		"strategy": "NAIVE",
	  
		"_methodLimitationEnabled": "limit or not",
		"methodLimitationEnabled": true,
	  
		"_allowedMethods": "can be ignored when set methodLimitationEnabled false",
		"allowedMethods": ["eth_blockNumber"],
	  
		"_contractWhitelist": "can be ignored when set methodLimitationEnabled false",
		"contractWhitelist": []
	  }`

	ctx := context.Background()

	config := &Config{}

	err := json.Unmarshal([]byte(testConfigStr1), config)

	currentRunningConfig, err = BuildRunningConfigFromConfig(ctx, config)

	if err != nil {
		logrus.Fatal(err)
	}

	assert.Equal(t, true, isAllowedMethod("eth_blockNumber"))
	assert.Equal(t, false, isAllowedMethod("eth_getBalance"))
}

func TestInWhitelist(t *testing.T) {
	var testConfigStr2 = `{
		"_upstreams": "support http, https, ws, wss",
		"upstreams": [
		  "https://ropsten.infura.io/v3/83438c4dcf834ceb8944162688749707"
		],
	  
		"_strategy": "support NAIVE, RACE, FALLBACK",
		"strategy": "NAIVE",
	  
		"_methodLimitationEnabled": "limit or not",
		"methodLimitationEnabled": true,
	  
		"_allowedMethods": "can be ignored when set methodLimitationEnabled false",
		"allowedMethods": ["eth_blockNumber", "eth_getBalance", "eth_call", "eth_sendRawTransaction"],
	  
		"_contractWhitelist": "can be ignored when set methodLimitationEnabled false",
		"contractWhitelist": ["0xc2c57336e01695D34F8012f6c0d250baB2Dd38Da"]
	  }`

	ctx := context.Background()

	config := &Config{}

	err := json.Unmarshal([]byte(testConfigStr2), config)

	currentRunningConfig, err = BuildRunningConfigFromConfig(ctx, config)

	if err != nil {
		logrus.Fatal(err)
	}

	assert.Equal(t, true, inWhitelist("0xc2c57336e01695D34F8012f6c0d250baB2Dd38Da"))
	assert.Equal(t, false, inWhitelist("0x126aa4Ef50A6e546Aa5ecD1EB83C060fB780891a"))
}

func TestIsValidCall(t *testing.T) {
	var testConfigStr2 = `{
		"_upstreams": "support http, https, ws, wss",
		"upstreams": [
		  "https://ropsten.infura.io/v3/83438c4dcf834ceb8944162688749707"
		],
	  
		"_strategy": "support NAIVE, RACE, FALLBACK",
		"strategy": "NAIVE",
	  
		"_methodLimitationEnabled": "limit or not",
		"methodLimitationEnabled": true,
	  
		"_allowedMethods": "can be ignored when set methodLimitationEnabled false",
		"allowedMethods": ["eth_blockNumber", "eth_getBalance", "eth_call", "eth_sendRawTransaction"],
	  
		"_contractWhitelist": "can be ignored when set methodLimitationEnabled false",
		"contractWhitelist": ["0xc2c57336e01695D34F8012f6c0d250baB2Dd38Da"]
	  }`

	ctx := context.Background()

	config := &Config{}

	err := json.Unmarshal([]byte(testConfigStr2), config)

	currentRunningConfig, err = BuildRunningConfigFromConfig(ctx, config)

	if err != nil {
		logrus.Fatal(err)
	}

	requestData1 := &RequestData{
		JsonRpc: "2.0",
		ID:      1,
		Method:  "eth_blockNumber",
		Params:  nil,
	}

	assert.Equal(t, nil, isValidCall(requestData1))

	requestData2 := &RequestData{
		JsonRpc: "2.0",
		ID:      1,
		Method:  "eth_getBalance",
		Params:  nil,
	}

	assert.Equal(t, nil, isValidCall(requestData2))

	requestData3 := &RequestData{
		JsonRpc: "2.0",
		ID:      1,
		Method:  "eth_call",
		Params:  nil,
	}

	assert.Equal(t, DecodeError, isValidCall(requestData3))

	requestData4 := &RequestData{
		JsonRpc: "2.0",
		ID:      1,
		Method:  "eth_sendRawTransaction",
		Params:  nil,
	}

	assert.Equal(t, DecodeError, isValidCall(requestData4))

	requestData5 := &RequestData{
		JsonRpc: "2.0",
		ID:      1,
		Method:  "eth_blockNumber_test",
		Params:  nil,
	}

	assert.Equal(t, DeniedMethod, isValidCall(requestData5))

	requestData6 := &RequestData{
		JsonRpc: "2.0",
		ID:      1,
		Method:  "eth_call",
		Params:  []interface{}{map[string]interface{}{"to": "0xc2c57336e01695D34F8012f6c0d250baB2Dd38Dd"}},
	}

	// to := requestData6.Params[0].(map[string]interface{})["to"].(string)
	assert.Equal(t, DeniedContract, isValidCall(requestData6))

	requestData7 := &RequestData{
		JsonRpc: "2.0",
		ID:      1,
		Method:  "eth_call",
		Params:  []interface{}{map[string]interface{}{"to": "0xc2c57336e01695D34F8012f6c0d250baB2Dd38Da"}},
	}

	assert.Equal(t, nil, isValidCall(requestData7))
}
