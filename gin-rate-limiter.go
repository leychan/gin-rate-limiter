package gin_rate_limiter

import (
	"github.com/redis/go-redis/v9"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	ratelimiter "github.com/leychan/go-rate-limiter"
)

var CommonKeyPrefix = "request:ratelimter:"
var redisClient *redis.Client
var RedisOpt *redis.Options

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
		return c.Request.URL.Path + getRequestClientIp(c)
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
	if c.GetHeader("X-Request-ID") != "" {
		return c.GetHeader("X-Request-ID")
	}

	if c.GetHeader("Request-ID") != "" {
		return c.GetHeader("Request-ID")
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
	if redisClient != nil {
		return redisClient
	}
	return newRedisClient()
}

// newRedisClient 创建redis客户端
func newRedisClient() *redis.Client {
	return redis.NewClient(RedisOpt)
}
