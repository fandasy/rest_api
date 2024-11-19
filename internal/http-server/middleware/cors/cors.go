package cors

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Middleware() gin.HandlerFunc {
	fn := func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")                   // Позволяет доступ с любого источника
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS") // Разрешенные методы
		c.Header("Access-Control-Allow-Headers", "Content-Type")       // Разрешенные заголовки

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}

	return fn
}
