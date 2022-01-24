package core

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func IsSignatureValid(signature string, containerKey string, expiration int) bool {
	h := hmac.New(sha256.New, []byte(_config.SignSecret))

	payload := fmt.Sprintf("%s-%d", containerKey, expiration)

	h.Write([]byte(payload))

	sha := hex.EncodeToString(h.Sum(nil))

	return sha == signature
}

func IsAdminToken(token string) bool {
	return _config.AdminToken == token
}
