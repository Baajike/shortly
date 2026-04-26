package utils

import (
	"crypto/rand"
	"math/big"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// GenerateNanoID returns a cryptographically random string of the given length
// drawn from the alphanumeric alphabet [0-9a-zA-Z].
func GenerateNanoID(length int) (string, error) {
	alphabetLen := big.NewInt(int64(len(alphabet)))
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			return "", err
		}
		result[i] = alphabet[n.Int64()]
	}
	return string(result), nil
}
