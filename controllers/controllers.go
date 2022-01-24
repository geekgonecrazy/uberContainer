package controllers

import (
	"strconv"
	"strings"
	"time"

	"github.com/geekgonecrazy/uberContainer/core"
	"github.com/gin-gonic/gin"
)

func isSignatureValid(containerKey string, c *gin.Context) bool {
	signature := c.GetHeader("X-Uber-Signature")
	exprHeader := c.GetHeader("X-Uber-Signature-Expire")

	if signature == "" {
		return false
	}

	expiration, err := strconv.Atoi(exprHeader)
	if err != nil {
		return false
	}

	expirationDate := time.Unix(int64(expiration), 0)

	if expirationDate.Before(time.Now()) {
		return false
	}

	if !core.IsSignatureValid(signature, containerKey, expiration) {
		return false
	}

	return true
}

func hasAdminToken(c *gin.Context) bool {
	bearer := c.GetHeader("Authorization")

	parts := strings.Split(bearer, " ")

	if len(parts) != 2 {
		return false
	}

	token := parts[1]

	return core.IsAdminToken(token)
}

func checkValidAuthentication(containerKey string, c *gin.Context) bool {
	if hasAdminToken(c) {
		return true
	}

	if isSignatureValid(containerKey, c) {
		return true
	}

	return false
}
