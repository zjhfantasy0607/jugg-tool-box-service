package common

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

const (
	HeartbeatInterval = 10 * time.Second // 心跳检测时间间隔
	ReadTimeout       = 60 * time.Second // 连接读取超时
)

var upgrader = websocket.Upgrader{
	// 允许所有 CORS 跨域请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func CreateWS(c *gin.Context) (*websocket.Conn, error) {
	// 将 HTTP 连接升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("升级连接失败: %w", err))
	}

	conn.SetReadDeadline(time.Now().Add(ReadTimeout))
	conn.SetPongHandler(func(appData string) error {
		fmt.Println("pong")
		conn.SetReadDeadline(time.Now().Add(ReadTimeout))
		return nil
	})

	// 启动心跳检测协程
	go func() {
		ticker := time.NewTicker(HeartbeatInterval)
		defer ticker.Stop()

		for range ticker.C {
			// 发送 ping 消息
			fmt.Println("ping")
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	return conn, nil
}
