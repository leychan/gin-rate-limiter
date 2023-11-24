package gin_rate_limiter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redis/go-redis/v9"

	"github.com/gin-gonic/gin"
)

// TestGetRequestClientIp 测试获取请求的客户端 IP
func TestGetRequestClientIp(t *testing.T) {
	// 创建一个 Gin 引擎
	r := gin.Default()
	r.GET("/test", func(c *gin.Context) {
		ip := getRequestClientIp(c)
		c.String(http.StatusOK, ip)
	})

	// 创建一个模拟的 HTTP 请求
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// 设置请求的客户端 IP
	req.Header.Set("X-Forwarded-For", "192.168.0.1")

	// 创建一个 ResponseRecorder 来记录响应
	rr := httptest.NewRecorder()

	// 将请求发送到 Gin 引擎进行处理
	r.ServeHTTP(rr, req)

	// 检查响应的状态码是否为 200
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rr.Code)
	}

	// 检查响应的主体是否与预期相符
	expected := "192.168.0.1"
	if rr.Body.String() != expected {
		t.Errorf("Expected response body %s, but got %s", expected, rr.Body.String())
	}
}

// TestGetRequestId 测试获取请求的唯一 ID
func TestGetRequestId(t *testing.T) {
	// Test case 1: X-Request-ID header is present
	req := &http.Request{}
	req.Header = make(http.Header)

	c := gin.Context{
		Request: req,
	}
	c.Request.Header.Set("X-Request-ID", "abc123")
	expected1 := "abc123"
	if got := getRequestId(&c); got != expected1 {
		t.Errorf("getRequestId() = %v, want %v", got, expected1)
	}

	// Test case 2: Request-ID header is present
	c.Request.Header.Set("X-Request-ID", "def456")
	expected2 := "def456"
	if got := getRequestId(&c); got != expected2 {
		t.Errorf("getRequestId() = %v, want %v", got, expected2)
	}

	// Test case 3: Both X-Request-ID and Request-ID headers are empty
	// c.Request.Header = http.Header{}
	// expected3 := "" // UUID string generated
	// if got := getRequestId(&c); got == "" || len(got) != 36 {
	//     t.Errorf("getRequestId() = %v, want a UUID string", got)
	// }
}

// TestApiSingleIpRateLimiterMiddleware 测试单个 API 和 IP 的限流
func TestApiSingleIpRateLimiterMiddleware(t *testing.T) {
	r := gin.Default()
	r.Use(ApiSingleIpRateLimiterMiddleware(20000, 2, func(ctx *gin.Context) {
		ctx.AbortWithStatus(http.StatusTooManyRequests)
	}))
	r.GET("/api/v1/users", func(c *gin.Context) {
		ip := getRequestClientIp(c)
		c.String(http.StatusOK, ip)
	})
	RedisOpt = &redis.Options{Addr: "127.0.0.1:6379"}
	// Test case 1: Request with different API paths and IP addresses
	t.Run("Different API paths and IP addresses", func(t *testing.T) {
		// Set up test environment

		// Define test cases
		testCases := []struct {
			path       string
			ip         string
			statusCode int
		}{
			{"/api/v1/users?uid=1", "127.0.0.1:80", http.StatusOK},
			{"/api/v1/users?uid=1", "127.0.0.1:80", http.StatusOK},
			{"/api/v1/users?uid=2", "127.0.0.1:80", http.StatusTooManyRequests},
			{"/api/v1/users?uid=2", "127.0.0.2:80", http.StatusOK},
			{"/api/v1/users?uid=3", "127.0.0.2:80", http.StatusOK},
			{"/api/v1/users?uid=3", "127.0.0.2:80", http.StatusTooManyRequests},
		}

		// Execute test cases
		for _, tc := range testCases {
			req, err := http.NewRequest(http.MethodGet, tc.path, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.RemoteAddr = tc.ip

			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)

			if resp.Code != tc.statusCode {
				t.Errorf("expected status code %d, got %d", tc.statusCode, resp.Code)
			}
		}
	})
}

// TestApiRateLimiterMiddleware 测试单个 API 的限流
func TestApiRateLimiterMiddleware(t *testing.T) {
	r := gin.Default()
	r.Use(ApiRateLimiterMiddleware(20000, 2, func(ctx *gin.Context) {
		ctx.AbortWithStatus(http.StatusTooManyRequests)
	}))
	r.GET("/api/v1/users", func(c *gin.Context) {
		ip := getRequestClientIp(c)
		c.String(http.StatusOK, ip)
	})
	r.GET("/api/v2/users", func(c *gin.Context) {
		ip := getRequestClientIp(c)
		c.String(http.StatusOK, ip)
	})
	RedisOpt = &redis.Options{Addr: "127.0.0.1:6379"}
	// Test case 1: Request with different API paths and IP addresses
	t.Run("Different API paths", func(t *testing.T) {
		// Set up test environment

		// Define test cases
		testCases := []struct {
			path       string
			ip         string
			statusCode int
		}{
			{"/api/v1/users", "127.0.0.1:80", http.StatusOK},
			{"/api/v1/users", "127.0.0.1:80", http.StatusOK},
			{"/api/v1/users", "127.0.0.1:80", http.StatusTooManyRequests},
			{"/api/v2/users", "127.0.0.1:80", http.StatusOK},
			{"/api/v2/users", "127.0.0.1:80", http.StatusOK},
			{"/api/v2/users", "127.0.0.1:80", http.StatusTooManyRequests},
		}

		// Execute test cases
		for _, tc := range testCases {
			req, err := http.NewRequest(http.MethodGet, tc.path, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.RemoteAddr = tc.ip

			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)

			if resp.Code != tc.statusCode {
				t.Errorf("expected status code %d, got %d", tc.statusCode, resp.Code)
			}
		}
	})
}

// TestApiRateLimiterMiddleware 测试单个 API 的限流
func TestGlobalRateLimiterMiddleware(t *testing.T) {
	r := gin.Default()
	r.Use(GlobalRateLimiterMiddleware(20000, 5, func(ctx *gin.Context) {
		ctx.AbortWithStatus(http.StatusTooManyRequests)
	}))
	r.GET("/api/v1/users", func(c *gin.Context) {
		ip := getRequestClientIp(c)
		c.String(http.StatusOK, ip)
	})
	r.GET("/api/v2/users", func(c *gin.Context) {
		ip := getRequestClientIp(c)
		c.String(http.StatusOK, ip)
	})
	RedisOpt = &redis.Options{Addr: "127.0.0.1:6379"}
	// Test case 1: Request with different API paths and IP addresses
	t.Run("Global APIs", func(t *testing.T) {
		// Set up test environment

		// Define test cases
		testCases := []struct {
			path       string
			ip         string
			statusCode int
		}{
			{"/api/v1/users", "127.0.0.1:80", http.StatusOK},
			{"/api/v1/users", "127.0.0.1:80", http.StatusOK},
			{"/api/v1/users", "127.0.0.1:80", http.StatusOK},
			{"/api/v2/users", "127.0.0.1:80", http.StatusOK},
			{"/api/v2/users", "127.0.0.1:80", http.StatusOK},
			{"/api/v2/users", "127.0.0.1:80", http.StatusTooManyRequests},
		}

		// Execute test cases
		for _, tc := range testCases {
			req, err := http.NewRequest(http.MethodGet, tc.path, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.RemoteAddr = tc.ip

			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)

			if resp.Code != tc.statusCode {
				t.Errorf("expected status code %d, got %d", tc.statusCode, resp.Code)
			}
		}
	})
}
