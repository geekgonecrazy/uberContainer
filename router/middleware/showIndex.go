package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func ShowIndex(c *gin.Context) {
	if strings.Contains(c.Request.RequestURI, "/api/") {
		c.JSON(404, gin.H{"status": 404})
		return
	}

	c.HTML(200, "index.tmpl", gin.H{})
}
