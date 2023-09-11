package middleware

import (
	"log/slog"
	"math"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		uri := c.Request.URL.Path
		sTime := time.Now()

		c.Next()

		if uri == "/health" {
			return
		}

		eTime := time.Now()
		latency := int(math.Ceil(float64(eTime.Sub(sTime).Nanoseconds()) / 1000000.0))
		httpCode := c.Writer.Status()
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}

		fields := []any{
			slog.Int("dataLength", dataLength),
			slog.Int64("receiveTime", sTime.UnixNano()/int64(time.Millisecond)),
			slog.Int64("responseTime", eTime.UnixNano()/int64(time.Millisecond)),
			slog.Int("httpCode", httpCode),
			slog.Int("latency", latency),
			slog.String("method", c.Request.Method),
			slog.String("uri", uri),
			slog.String("clientIP", c.ClientIP()),
		}

		if len(c.Errors) > 0 {
			slog.Error(c.Errors.ByType(gin.ErrorTypePrivate).String(), fields)
		} else {
			switch {
			case httpCode > 499:
				slog.Error("request", fields...)
			case httpCode > 399:
				slog.Warn("request", fields...)
			default:
				slog.Info("request", fields...)
			}
		}
	}
}
