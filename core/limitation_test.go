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
		"contractWhitelist": ["0x06898143df04616a8a8f9614deb3b99ba12b3096"]
	  }`

	ctx := context.Background()

	config := &Config{}

	err := json.Unmarshal([]byte(testConfigStr2), config)

	currentRunningConfig, err = BuildRunningConfigFromConfig(ctx, config)

	if err != nil {
		logrus.Fatal(err)
	}

	assert.Equal(t, true, inWhitelist("0x06898143df04616a8a8f9614deb3b99ba12b3096"))
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
		"contractWhitelist": ["0x06898143df04616a8a8f9614deb3b99ba12b3096"]
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
		Params:  []interface{}{map[string]interface{}{"to": "0x06898143df04616a8a8f9614deb3b99ba12b3096"}},
	}

	assert.Equal(t, nil, isValidCall(requestData7))

	requestData8 := &RequestData{
		JsonRpc: "2.0",
		ID:      1,
		Method:  "eth_sendRawTransaction",
		Params:  []interface{}{`0xffffffffffffffffffffffffffffffffffff`},
	}

	assert.Equal(t, DecodeError, isValidCall(requestData8))

	requestData9 := &RequestData{
		JsonRpc: "2.0",
		ID:      1,
		Method:  "eth_sendRawTransaction",
		Params:  []interface{}{"0xf9018b14850306dc420083025db89406898143df04616a8a8f9614deb3b99ba12b309680b901248059cf3b000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000000300000000000000000000000060fa59b6a32c08023c5e0002d6ddebdf4cb2c294000000000000000000000000000000000000000000000000000000002a45d6a02aa0a400038e05162401a612414b0129b7a0fab2824fdb7d365a4e9c34309b633aa5a02cd68de2b4146542a4fed0d918d011617e75d84f024dee4b0028dff56e1f9b31"},
	}

	assert.Equal(t, nil, isValidCall(requestData9))

	requestData10 := &RequestData{
		JsonRpc: "2.0",
		ID:      1,
		Method:  "eth_sendRawTransaction",
		Params:  []interface{}{"0xf9018b14850306dc420083025db89406898143df04616a8a8f9014deb3b99ba12b309680b901248059cf3b000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000000300000000000000000000000060fa59b6a32c08023c5e0002d6ddebdf4cb2c294000000000000000000000000000000000000000000000000000000002a45d6a02aa0a400038e05162401a612414b0129b7a0fab2824fdb7d365a4e9c34309b633aa5a02cd68de2b4146542a4fed0d918d011617e75d84f024dee4b0028dff56e1f9b31"},
	}

	assert.Equal(t, DeniedContract, isValidCall(requestData10))
}
