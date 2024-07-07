package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
	"sync"
)

type Server struct {
	sync.RWMutex // 读写锁

	opt            *option
	authentication Authentication

	routes map[string]HandlerFunc // 方法名和handler对应的路由映射
	addr   string

	connToUser map[*Conn]string
	userToConn map[string]*Conn

	upgrader websocket.Upgrader // 将http升级为WebSocket连接
	logx.Logger
}

func NewServer(addr string, opts ...Options) *Server {
	opt := newOption(opts...)

	return &Server{
		authentication: opt.Authentication,
		opt:            &opt,
		addr:           addr,
		upgrader:       websocket.Upgrader{},
		Logger:         logx.WithContext(context.Background()),

		routes:     make(map[string]HandlerFunc),
		connToUser: make(map[*Conn]string),
		userToConn: make(map[string]*Conn),
	}
}

func (s *Server) ServerWs(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			s.Errorf("server handler ws recover err %v", err)
		}
	}()

	// 对Conn通道进行了封装
	conn := NewConn(s, w, r)
	if conn == nil {
		return
	}

	// 判断该请求是否有访问该服务器的权限
	if !s.authentication.Auth(w, r) {
		s.Send(&Message{Type: FrameErr, Data: fmt.Sprintf("不具备访问权限")}, conn)
		conn.Close()
		return
	}

	// 添加连接记录
	s.addConn(conn, r)

	// 处理连接
	go s.handlerConn(conn)
}

// 会有并发问题，需要加锁
func (s *Server) addConn(conn *Conn, req *http.Request) {
	uid := s.authentication.UserId(req)

	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()

	// 验证用户之前是否登陆过
	if c := s.userToConn[uid]; c != nil {
		// 关闭之前的连接
		c.Close()
	}

	s.connToUser[conn] = uid
	s.userToConn[uid] = conn
}

func (s *Server) handlerConn(conn *Conn) {
	// 记录连接
	for { // 无限循环
		// 获取请求消息
		_, msg, err := conn.ReadMessage()
		if err != nil {
			s.Errorf("websocket conn readMessage err %v, user Id %s", err, "")
			// 关闭并删除连接
			s.Close(conn)
			return
		}

		// 请求信息
		var message Message
		if err = json.Unmarshal(msg, &message); err != nil {
			fmt.Println(json.Unmarshal(msg, &message))
			s.Send(NewErrorMessage(err), conn)
			s.Close(conn)
			return
		}

		// 依据请求消息类型分类处理
		switch message.Type {
		case FramePing:
			// ping
			s.Send(&Message{Type: FramePing}, conn)
		case FrameData:
			// 处理
			if handler, ok := s.routes[message.Method]; ok {
				handler(s, conn, &message)
			} else {
				s.Send(&Message{
					Type: FrameErr,
					Data: fmt.Sprintf("不存在请求方法 %v 请仔细检查", message.Method),
				}, conn)
			}
		}
	}
}

func (s *Server) SendByUserId(msg interface{}, sendIds ...string) error {
	if len(sendIds) == 0 {
		return nil
	}

	return s.Send(msg, s.GetConns(sendIds...)...)
}

func (s *Server) Send(msg interface{}, conns ...*Conn) error {
	if len(conns) == 0 {
		return nil
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	for _, conn := range conns {
		if err = conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) AddRoute(rs []Route) {
	for _, r := range rs {
		s.routes[r.Method] = r.Handler
	}
}

func (s *Server) Start() {
	http.HandleFunc(s.opt.pattern, s.ServerWs)
	http.ListenAndServe(s.addr, nil)
}

func (s *Server) Stop() {
	fmt.Println("stop server")
}

func (s *Server) GetConn(uid string) *Conn {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	return s.userToConn[uid]
}

func (s *Server) GetConns(uids ...string) []*Conn {
	if len(uids) == 0 {
		return nil
	}

	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	res := make([]*Conn, 0, len(uids))
	for _, uid := range uids {
		res = append(res, s.userToConn[uid])
	}

	return res
}

func (s *Server) GetUsers(conns ...*Conn) []string {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	var res []string
	if len(conns) == 0 {
		// 获取全部
		res = make([]string, 0, len(s.connToUser))
		for _, uid := range s.connToUser {
			res = append(res, uid)
		}
	} else {
		// 获取部分
		res = make([]string, 0, len(conns))
		for _, conn := range conns {
			res = append(res, s.connToUser[conn])
		}
	}
	return res
}

func (s *Server) Close(conn *Conn) {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()

	uid := s.connToUser[conn]
	if uid == "" {
		// 已经关闭了连接
		return
	}

	delete(s.connToUser, conn)
	delete(s.userToConn, uid)

	conn.Close()
}
