package pkg

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func NewLogger() *slog.Logger {
	// JSON handler sur stdout
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, // LevelDebug si tu veux plus de détails
	})
	return slog.New(h)
}

func Logging(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		dur := time.Since(start)

		reqID, _ := c.Get("request_id")
		logger.Info("http_request",
			slog.String("id", toString(reqID)),
			slog.String("method", c.Request.Method),
			slog.String("route", c.FullPath()),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Int("bytes", c.Writer.Size()),
			slog.Duration("latency", dur),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}
