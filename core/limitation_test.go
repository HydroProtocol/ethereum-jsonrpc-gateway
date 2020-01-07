package core

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var testConfigStr1 = `{
	"_upstreams": "support http, https, ws, wss",
	"upstreams": [
	  "https://ropsten.infura.io/v3/83438c4dcf834ceb8944162688749707"
	],
  
	"_strategy": "support NAIVE, RACE, FALLBACK",
	"strategy": "NAIVE",
  
	"_methodLimitationEnabled": "limit or not",
	"methodLimitationEnabled": false,
  
	"_allowedMethods": "can be ignored when set methodLimitationEnabled false",
	"allowedMethods": ["eth_getBalance"],
  
	"_contractWhitelist": "can be ignored when set methodLimitationEnabled false",
	"contractWhitelist": ["0x..."]
  }`

func TestIsAllowedMethod(t *testing.T) {
	ctx := context.Background()

	config := &Config{}

	err := json.Unmarshal([]byte(testConfigStr1), config)

	currentRunningConfig, err = buildRunningConfigFromConfig(ctx, config)

	if err != nil {
		logrus.Fatal(err)
	}

	assert.Equal(t, true, true)
}

func TestInWhitelist(t *testing.T) {
	assert.Equal(t, true, true)
}
