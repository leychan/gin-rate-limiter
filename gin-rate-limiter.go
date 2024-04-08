package gin_rate_limiter

import (
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	ratelimiter "github.com/leychan/go-rate-limiter"
)

var CommonKeyPrefix = "request:ratelimter:"
var redisClient *redis.Client
var RedisOpt *redis.Options
var redisClientMu sync.Mutex

// RateLimiterMiddleware returns a Gin middleware that implements rate limiting for a single API
func RateLimiterMiddleware(
	windowSize int64, threshold int64,
	abort func(ctx *gin.Context),
	getKey func(*gin.Context) string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		// redis key
		key := CommonKeyPrefix + getKey(c)
		// unique id
		requestId := getRequestId(c)
		swOpt := ratelimiter.NewSlideWindowOpt(windowSize, threshold, ratelimiter.WithKey(key), ratelimiter.WithUniqueId(requestId))
		if !ratelimiter.NewLimiter(swOpt, getRedisClient()).CheckLimited() {
			abort(c)
			return
		}
		c.Next()
	}
}

// ApiRateLimiterMiddleware returns a Gin middleware that implements rate limiting for a single API
func ApiRateLimiterMiddleware(windowSize int64, threshold int64, abort func(ctx *gin.Context)) gin.HandlerFunc {
	return RateLimiterMiddleware(windowSize, threshold, abort, func(ctx *gin.Context) string {
		return ctx.Request.URL.Path
	})
}

// ApiSingleIpRateLimiterMiddleware returns a Gin middleware that implements rate limiting for a single API and IP
func ApiSingleIpRateLimiterMiddleware(windowSize int64, threshold int64, abort func(ctx *gin.Context)) gin.HandlerFunc {
	return RateLimiterMiddleware(windowSize, threshold, abort, func(c *gin.Context) string {
		var builder strings.Builder
		builder.WriteString(getRequestClientIp(c))
		builder.WriteString(":")
		builder.WriteString(c.Request.URL.Path)		
		return builder.String()
	})
}

// GlobalRateLimiterMiddleware returns a Gin middleware that implements rate limiting for all requests
func GlobalRateLimiterMiddleware(windowSize int64, threshold int64, abort func(ctx *gin.Context)) gin.HandlerFunc {
	return RateLimiterMiddleware(windowSize, threshold, abort, func(c *gin.Context) string {
		return "global"
	})
}

// GetRequestID 获取请求id,如果没有则生成一个
func getRequestId(c *gin.Context) string {
	requestID := c.GetHeader("X-Request-ID")
	if requestID != "" {
		return requestID
	}
	requestID = c.GetHeader("Request-ID")
	if requestID != "" {
		return requestID
	}

	return uuid.NewString()
}

// GetRequestClientIp 获取请求的客户端ip
func getRequestClientIp(c *gin.Context) string {
	ip := strings.Split(c.ClientIP(), ",")[0]
	return strings.TrimSpace(ip)
}

// getRedisClient 获取redis客户端
func getRedisClient() *redis.Client {
	redisClientMu.Lock()
	if redisClient == nil {
		redisClient = redis.NewClient(RedisOpt)
	}
	redisClientMu.Unlock()
	return redisClient
}

// SetRedisClient 设置redis客户端
func SetRedisClient(client *redis.Client) {
	redisClientMu.Lock()
	redisClient = client
	redisClientMu.Unlock()
}