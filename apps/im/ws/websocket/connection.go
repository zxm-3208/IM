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

	messageMu      sync.Mutex
	readMessages   []*Message          // 接收信息处理队列
	readMessageSeq map[string]*Message // 用于ACk验证，记录ACK机制中消息的处理结果和进展
	message        chan *Message       // 消息通道，在ack验证完成后将消息投递到writeHandler进行处理
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
		message:           make(chan *Message, 1), //1保证可以缓存一个Message的值，且保证了发送和接收操作的顺序，从而确保了数据的有序性
		readMessageSeq:    make(map[string]*Message, 2),
		readMessages:      make([]*Message, 0, 2),
	}

	go conn.keepalive()
	return conn
}

// 将消息记录到队列中
func (c *Conn) appendMsgMq(msg *Message) {
	c.messageMu.Lock()
	defer c.messageMu.Unlock()

	// 读队列 (验证之前是否存在ack记录)
	if m, ok := c.readMessageSeq[msg.Id]; ok {
		// 已经有消息的记录，该消息已经有ack的确认(重复发送，或者收到了ack消息，已被处理了)
		if len(c.readMessages) == 0 {
			return
		}
		// 重复消息 (已经发送消息)
		if m.Id != msg.Id || m.AckSeq >= msg.AckSeq {
			return
		}
		// 更新最新ack记录
		c.readMessageSeq[msg.Id] = msg
		return
	}
	// 还没有进行ack的确认, 避免客户端重复发送多余的ack消息
	if msg.Type == FrameAck {
		return
	}

	c.readMessages = append(c.readMessages, msg)
	c.readMessageSeq[msg.Id] = msg
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
	c.idleMu.Lock()
	defer c.idleMu.Unlock()
	messageType, p, err = c.Conn.ReadMessage() // 平时阻塞，只有有数据到来了才执行
	c.idle = time.Time{}                       // 零值的 time.Time 实例
	return
}

func (c *Conn) WriteMessage(messageType int, data []byte) error {
	c.idleMu.Lock() // 防止并发写错误
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
