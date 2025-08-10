package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("正常系: レートリミットに達しない場合", func(t *testing.T) {
		r := gin.New()
		r.Use(RateLimitMiddleware(5, time.Minute))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		for i := 0; i < 5; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("異常系: レートリミットに達した場合", func(t *testing.T) {
		r := gin.New()
		r.Use(RateLimitMiddleware(2, time.Minute))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// 最初の2回はOK
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}

		// 3回目でレートリミットに達する
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})

	t.Run("正常系: ウィンドウ期間が過ぎるとリセットされる", func(t *testing.T) {
		r := gin.New()
		r.Use(RateLimitMiddleware(1, 100*time.Millisecond))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// 1回目はOK
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// 期間内に2回目 -> Too Many Requests
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		// 期間が過ぎるのを待つ
		time.Sleep(150 * time.Millisecond)

		// 期間が過ぎたので再度OK
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
