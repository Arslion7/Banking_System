package rand

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func Bytes(n int) ([]byte, error) {
	b := make([]byte, n)
	num, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("Error in generating random")
	}
	if num < n {
		return nil, fmt.Errorf("Error in generating random")
	}
	return b, nil
}

func String(n int) (string, error) {
	b, err := Bytes(n)
	if err != nil {
		return "", fmt.Errorf("Error in generating random %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}