package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandStringRunes(t *testing.T) {
	str1 := RandStringRunes(8)
	assert.IsType(t, "string", str1)
	assert.Equal(t, 8, len(str1))

	str2 := RandStringRunes(64)
	assert.IsType(t, "string", str2)
	assert.Equal(t, 64, len(str2))
}

func TestNoErrorFieldInJSON(t *testing.T) {
	jsonStr1 := "{\"foo\": 100, \"bar\": \"Go\"}"
	result1 := NoErrorFieldInJSON(jsonStr1)
	assert.Equal(t, true, result1)

	jsonStr2 := "foobar"
	result2 := NoErrorFieldInJSON(jsonStr2)
	assert.Equal(t, false, result2)

	jsonStr3 := "{\"foo\": 100, \"error\": \"Go\"}"
	result3 := NoErrorFieldInJSON(jsonStr3)
	assert.Equal(t, false, result3)
}
