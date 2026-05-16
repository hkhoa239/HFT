package middleware

import "github.com/gin-gonic/gin"

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow both dev (4200) and production/docker (3000) ports
		origin := c.Request.Header.Get("Origin")
		if origin == "http://localhost:3000" || origin == "http://localhost:4200" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// Fallback for non-browser requests or other origins (consider tightening for production)
			c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
