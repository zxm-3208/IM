package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	"net/http"
	"sync"
	"time"
)

type AckType int

const (
	// 不进行ack确认
	NoAck AckType = iota
	// 只有一次回复的ack确认 - 两次通信
	OnlyAck
	// 严格的三次握手机制
	RigorAck
)

func (t AckType) ToString() string {
	switch t {
	case OnlyAck:
		return "OnlyAck"
	case RigorAck:
		return "RigorAck"
	}
	return "NoAck"
}

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

	*threading.TaskRunner
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
		TaskRunner: threading.NewTaskRunner(opt.concurrency), // 创建一个任务运行器，用于异步地执行任务,并指定了并发执行任务的最大数量
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

// 获取任务
func (s *Server) handlerConn(conn *Conn) {

	uids := s.GetUsers(conn)
	conn.Uid = uids[0]

	// 处理任务
	go s.handlerWrite(conn)

	if s.isAck(nil) {
		// 接收Ack确认
		go s.readAck(conn)
	}

	// 记录连接
	for {
		// 获取请求消息
		_, msg, err := conn.ReadMessage()
		fmt.Println("new msg ", string(msg), err)
		if err != nil {
			s.Errorf("websocket conn readMessage err %v, user Id %s", err, "")
			// 关闭并删除连接
			s.Close(conn)
			return
		}

		// 请求信息解析
		var message Message
		if err = json.Unmarshal(msg, &message); err != nil {
			fmt.Println(json.Unmarshal(msg, &message))
			s.Send(NewErrorMessage(err), conn)
			s.Close(conn)
			return
		}

		// 依据消息进行处理
		if s.isAck(&message) {
			s.Infof("conn message read ack msg %v", message)
			conn.appendMsgMq(&message) // 将消息记录到队列中, 发送过程在readAck方法中，当消息确认后发送到chan
		} else {
			conn.message <- &message
		}
	}
}

// 任务的处理
func (s *Server) handlerWrite(conn *Conn) {
	for {
		select {
		case <-conn.done:
			// 连接关闭
			return
		case message := <-conn.message:
			// 依据请求消息类型分类处理
			switch message.Type {
			case FramePing:
				s.Send(&Message{Type: FramePing}, conn)
			case FrameData:
				// 根据请求的method分发路由并执行
				if handler, ok := s.routes[message.Method]; ok {
					handler(s, conn, message)
				} else {
					s.Send(&Message{
						Type: FrameErr,
						Data: fmt.Sprintf("不存在请求方法 %v 请仔细检查", message.Method),
					}, conn)
				}
			}

			if s.isAck(message) {
				conn.messageMu.Lock()
				delete(conn.readMessageSeq, message.Id) // 已经完成，删除该消息的记录(1. 接收；2. 发送，在这里已经完成了发送，所以该消息的使命已经结束)
				conn.messageMu.Unlock()
			}
		}
	}
}

// 读取消息的ack
func (s *Server) readAck(conn *Conn) {

	send := func(msg *Message, conn *Conn) error {
		err := s.Send(msg, conn)
		if err == nil {
			return nil
		}
		s.Errorf("message ack OnlyAck send err %v message %v", err, msg)
		conn.readMessages[0].errCount++
		conn.messageMu.Unlock()
		// 随着重试次数增加，等待时间延长
		tempDelay := time.Duration(200*conn.readMessages[0].errCount) * time.Millisecond
		if max := 1 * time.Second; tempDelay > max {
			tempDelay = max
		}
		time.Sleep(tempDelay)
		return err
	}

	for {
		// 死循环的退出机制
		select {
		case <-conn.done:
			s.Infof("close message ack uid %v", conn.Uid)
			return
		default: // 没有任何通道准备就绪，select语句会再次阻塞，等待通道变成可通信状态
		}

		// 从队列中读取新的消息
		conn.messageMu.Lock() // 涉及到map与slice的读写处理, 需要加锁
		// 如果队列中没有消息了，就睡眠避免忙等
		if len(conn.readMessages) == 0 {
			conn.messageMu.Unlock()
			// 增加睡眠 (避免忙等，减少CPU使用率)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// 读取第一条消息
		message := conn.readMessages[0]
		// 判断是否超过重试次数
		if message.errCount > s.opt.sendErrCount {
			s.Infof("conn send fail, message %v, ackType %v, maxSendErrCount %v", message, s.opt.ack.ToString(), s.opt.sendErrCount)
			conn.messageMu.Unlock()
			// 因为发送消息多次错误，而选择放弃消息
			delete(conn.readMessageSeq, message.Id)
			conn.readMessages = conn.readMessages[1:]
			continue
		}

		// 判断ack的方式
		switch s.opt.ack {
		case OnlyAck:
			// 直接给客户端回复
			if err := send(&Message{
				Type:   FrameAck,
				Id:     message.Id,
				AckSeq: message.AckSeq + 1,
			}, conn); err != nil {
				continue
			}
			// 把消息从队列中移除
			conn.readMessages = conn.readMessages[1:]
			conn.messageMu.Unlock()
			conn.message <- message
			s.Infof("message ack OnlyAck send success mid %v", message.Id)
		case RigorAck:
			// 还未发送过确认信息
			if message.AckSeq == 0 {
				conn.readMessages[0].AckSeq++
				conn.readMessages[0].ackTime = time.Now()
				if err := send(&Message{
					Type:   FrameAck,
					AckSeq: message.AckSeq,
					Id:     message.Id,
				}, conn); err != nil {
					continue
				}
				conn.messageMu.Unlock()
				s.Infof("message ack RigorAck send mid %v, seq %v , time%v", message.Id, message.AckSeq, message.ackTime)
				continue
			}

			// 二次验证
			// 1. 客户端返回结果，再一次确认
			msgSeq := conn.readMessageSeq[message.Id]
			if msgSeq.AckSeq > message.AckSeq {
				// 客户端已确认
				conn.readMessages = conn.readMessages[1:]
				conn.messageMu.Unlock()
				conn.message <- message
				s.Infof("message ack RigorAck success mid %v", message.Id)
				continue
			}

			// 2. 客户端没有确认，考虑是否超过了ack的确认时间
			val := s.opt.ackTimeout - time.Since(message.ackTime)
			if !message.ackTime.IsZero() && val <= 0 {
				// 2.1 超时, 删除
				delete(conn.readMessageSeq, message.Id)
				conn.readMessages = conn.readMessages[1:]
				conn.messageMu.Unlock()
				s.Errorf("message ack RigorAck fail mid %v, time %v because timeout", message.Id, message.ackTime)
				continue
			}
			// 2.2 未超时，重新发送
			conn.messageMu.Unlock()
			if err := send(&Message{
				Type:   FrameAck,
				Id:     message.Id,
				AckSeq: message.AckSeq,
			}, conn); err != nil {
				continue
			}
			// 睡眠
			time.Sleep(3 * time.Second)
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

func (s *Server) isAck(message *Message) bool {
	if message == nil {
		return s.opt.ack != NoAck
	}
	return s.opt.ack != NoAck && message.Type != FrameNoAck
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
