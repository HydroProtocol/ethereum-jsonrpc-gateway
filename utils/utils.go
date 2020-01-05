package utils

import (
	"encoding/json"
	"log"
	"math/rand"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func NoErrorFieldInJSON(jsonStr string) bool {
	var tmp map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &tmp)

	if err != nil {
		log.Printf("decode json string failed, %v, %v\n", jsonStr, err)
		return false
	}

	if tmp["error"] == nil {
		return true
	}

	return false
}
