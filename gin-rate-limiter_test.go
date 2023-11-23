package gin_rate_limiter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

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
