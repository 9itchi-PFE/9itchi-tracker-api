package router

import (
	"crypto/subtle"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"9itchi-tracker-api/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func initializeMiddlewares(r *gin.Engine) {
	r.Use(cors())
	r.Use(RecoveryWithZap(utils.Logger, true))
	r.Use(Ginzap(utils.Logger, time.RFC3339, true, false))
	// dump request in debug
	if gin.Mode() == gin.DebugMode {
		r.Use(requestLogger())
	}
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Referrer, User-Agent, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// requestLogger logs method, path, status and latency only; never bodies.
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		utils.Logger.Logger.Debug("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}

func Ginzap(logger utils.LogWrapperObj, timeFormat string, utc bool, alwaysLog bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		if utc {
			end = end.UTC()
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				logger.Error(e)
			}
		} else if alwaysLog || c.Writer.Status() >= 300 {
			logger.Info(path,
				zap.Int("status", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", c.ClientIP()),
				zap.String("user-agent", c.Request.UserAgent()),
				zap.String("time", end.Format(timeFormat)),
				zap.Duration("latency", latency),
			)
		}
	}
}

func RecoveryWithZap(logger utils.LogWrapperObj, stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					logger.Error("[Recovery from panic]",
						zap.Time("time", time.Now()),
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					logger.Error("[Recovery from panic]",
						zap.Time("time", time.Now()),
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

// AuthMiddleware decrypts (AES-GCM) and verifies (HS256) the Bearer token, then sets
// user_id, email and role in the gin context. Only access tokens are accepted.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := utils.BearerToken(c.GetHeader("Authorization"))
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		claims, err := utils.ParseToken(utils.LoadConfig(), raw)
		if err != nil || claims.Type != utils.TokenAccess || claims.ID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Set("user_id", claims.ID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequireRole allows the request only when the authenticated role matches.
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") != role {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// InternalAPIKeyMiddleware guards service-to-service endpoints with a shared
// secret (env INTERNAL_API_KEY) sent in the x-api-key header. An unset key
// rejects every request. With INTERNAL_API_KEY_REQUIRED=false a bad or missing
// key is logged (never the key itself) and the request is allowed.
func InternalAPIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := utils.LoadConfig()
		provided := c.GetHeader("x-api-key")
		if cfg.InternalAPIKey != "" && subtle.ConstantTimeCompare([]byte(provided), []byte(cfg.InternalAPIKey)) == 1 {
			c.Next()
			return
		}
		if !cfg.InternalAPIKeyRequired {
			utils.Logger.Warn("internal_key_missing",
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()),
			)
			c.Next()
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
	}
}
