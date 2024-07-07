package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

type Conn struct {
	Uid string

	*websocket.Conn
	s *Server

	idleMu            sync.Mutex // 互斥锁
	idle              time.Time
	maxConnectionIdle time.Duration

	done chan struct{} //空结构体类型，它的大小为零。 对应的select用来非阻塞地监听一个或多个通道的活动。select语句会随机选择一个已经准备好的通道表达式进行读取或写入操作。
}

func NewConn(s *Server, w http.ResponseWriter, r *http.Request) *Conn {
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.Errorf("upgrade http conn err %v", err)
		return nil
	}

	conn := &Conn{
		Conn:              c,
		s:                 s,
		idle:              time.Now(),
		maxConnectionIdle: s.opt.maxConnectionIdle,
		done:              make(chan struct{}),
	}

	go conn.keepalive()
	return conn
}

// 长连接检测机制
func (c *Conn) keepalive() {
	idleTimer := time.NewTimer(c.maxConnectionIdle) // 定时器（检测是否超过最大空闲时间）
	defer idleTimer.Stop()

	for {
		select {
		case <-idleTimer.C:
			c.idleMu.Lock()
			idle := c.idle
			fmt.Printf("idle %v, maxIdle %v \n", c.idle, c.maxConnectionIdle)
			if idle.IsZero() { // 非空闲连接
				c.idleMu.Unlock()
				idleTimer.Reset(c.maxConnectionIdle) // 重置定时器时间
				continue
			}

			val := c.maxConnectionIdle - time.Since(idle) // time.Since用于计算从给定的时间点到当前时间的持续时间
			c.idleMu.Unlock()
			if val <= 0 {
				c.s.Close(c)
				return
			}
			idleTimer.Reset(val)
		case <-c.done:
			fmt.Println("客户端连接结束")
			return
		}
	}
}

func (c *Conn) ReadMessage() (messageType int, p []byte, err error) {
	// 开始忙碌
	messageType, p, err = c.Conn.ReadMessage() // 平时阻塞，只有有数据到来了才执行
	c.idleMu.Lock()
	defer c.idleMu.Unlock()
	c.idle = time.Time{} // 零值的 time.Time 实例
	return
}

func (c *Conn) WriteMessage(messageType int, data []byte) error {
	c.idleMu.Lock()
	defer c.idleMu.Unlock()
	err := c.Conn.WriteMessage(messageType, data) // 阻塞，直到数据被完全写入或遇到错误
	c.idle = time.Now()                           // 写入后空闲
	return err
}

func (c *Conn) Close() error {
	select {
	case <-c.done:
	default:
		close(c.done)
	}
	return c.Conn.Close()
}
