package websocket

import (
	"github.com/gorilla/websocket"
	"net/url"
)

/**
MQ中的websocker客户端，与ws中的ws服务端建立连接。在MQ中调用，当kafka异步处理完消息后，将msg发送给ws
*/

import (
	"encoding/json"
)

type Client interface {
	Close() error
	Send(v any) error
	Read(v any) error
}

type client struct {
	*websocket.Conn
	host string
	opt  dailOption
}

func NewClient(host string, opts ...DailOptions) *client {
	opt := newDailOption(opts...)
	c := client{
		Conn: nil,
		opt:  opt,
		host: host,
	}
	conn, err := c.dial()
	if err != nil {
		panic(err)
	}
	c.Conn = conn
	return &c
}

// 建立连接
func (c *client) dial() (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: c.host, Path: c.opt.pattern}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), c.opt.header)
	return conn, err
}

func (c *client) Send(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	err = c.WriteMessage(websocket.TextMessage, data)
	if err == nil {
		return nil
	}

	// 发送失败了再建立一次连接
	conn, err := c.dial()
	if err != nil {
		return err
	}

	c.Conn = conn
	return c.WriteMessage(websocket.TextMessage, data)
}

func (c *client) Read(v any) error {
	_, msg, err := c.Conn.ReadMessage()
	if err != nil {
		return err
	}
	return json.Unmarshal(msg, v)
}