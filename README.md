# gin-rate-limiter

gin-rate-limiter is a rate limiter for the gin framework, powered by redis.

## Installation

```bash
go get github.com/leychan/gin-rate-limiter
```

## Usage

```go
package main

import (
    "github.com/gin-gonic/gin"
    gin_rate_limiter "github.com/leychan/gin-rate-limiter"
    "github.com/redis/go-redis/v9"
)

// abort is a function that aborts the request with a 429 status code
var abort = func(c *gin.Context) {
    c.AbortWithStatusJSON(429, gin.H{"message": "too many requests"})
}

// apiRateLimiter returns a Gin middleware that implements rate limiting for a single API
func apiRateLimiter(windowSize int64, threshold int64) gin.HandlerFunc {
    return gin_rate_limiter.ApiRateLimiterMiddleware(windowSize, threshold, abort)
}

// apiSingleIpRateLimiter returns a Gin middleware that implements rate limiting for a single API and IP
func apiSingleIpRateLimiter(windowSize int64, threshold int64) gin.HandlerFunc {
    return gin_rate_limiter.ApiSingleIpRateLimiterMiddleware(windowSize, threshold, abort)
}

// globalRateLimiter returns a Gin middleware that implements rate limiting for all requests
func globalRateLimiter(windowSize int64, threshold int64) gin.HandlerFunc {
    return gin_rate_limiter.GlobalRateLimiterMiddleware(windowSize, threshold, abort)
}

func main() {
    r := gin.Default()
    // 限制每 100000 ms 时间内，所有服务最多 10 个请求
    r.Use(globalRateLimiter(100000, 10))

    //redis 配置
    gin_rate_limiter.RedisOpt = &redis.Options{Addr: "127.0.0.1:6379"}

    // 限制每 20000 ms 时间内，单个 API 最多 接受 5 个请求
    r.GET("/ping1", apiRateLimiter(20000, 5), func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })

    // 限制每 20000 ms 时间内，单个 API 和单个 IP 最多 接受 5 个请求
    r.GET("/ping2", apiSingleIpRateLimiter(20000, 5), func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })

    r.Run(":8888")
}

```


