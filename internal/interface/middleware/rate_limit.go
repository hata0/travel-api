package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type client struct {
	lastRequest time.Time
	requestCount int
}

// RateLimitMiddleware はIPアドレスベースのシンプルなレートリミットミドルウェアを返します。
// limit: 期間内の最大リクエスト数
// window: リクエスト数をカウントする期間
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	clients := make(map[string]*client)
	mu := &sync.Mutex{}

	// 古いエントリをクリーンアップするゴルーチン
	go func() {
		for {
			<-time.After(window)
			mu.Lock()
			for ip, cli := range clients {
				if time.Since(cli.lastRequest) > window {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		cli, found := clients[ip]
		if !found {
			cli = &client{lastRequest: time.Now(), requestCount: 0}
			clients[ip] = cli
		}

		// ウィンドウ期間が過ぎていたらリセット
		if time.Since(cli.lastRequest) > window {
			cli.requestCount = 0
			cli.lastRequest = time.Now()
		}

		cli.requestCount++

		if cli.requestCount > limit {
			mu.Unlock()
			slog.Warn("Rate limit exceeded", "ip", ip, "count", cli.requestCount)
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		mu.Unlock()

		c.Next()
	}
}
