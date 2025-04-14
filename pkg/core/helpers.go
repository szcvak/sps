package core

import (
	"math/rand/v2"
)

const (
	allowedCodeCharacters = "23456789ABCDEFGHJKLMNPQRSTUVWYZ"
)

func GenerateTeamCode() string {
	code := make([]byte, 6)

	for i := range code {
		code[i] = allowedCodeCharacters[rand.IntN(len(allowedCodeCharacters))]
	}

	return string(code)
}
