package util

import (
	"encoding/hex"
	"math/rand"
)

// Generate a random hash string in hexadecimal format.
func RandoHash(numCharacters int) (string, error) {
	b := make([]byte, numCharacters)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[0:numCharacters], nil
}
