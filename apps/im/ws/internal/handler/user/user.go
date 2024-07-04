package user

import (
	"IM/apps/im/ws/internal/svc"
	myWebsocket "IM/apps/im/ws/websocket"
)

// 获取在线的用户
func OnLine(svc *svc.ServiceContext) myWebsocket.HandlerFunc {
	return func(srv *myWebsocket.Server, conn *myWebsocket.Conn, msg *myWebsocket.Message) {
		uids := srv.GetUsers()
		u := srv.GetUsers(conn)
		err := srv.Send(myWebsocket.NewMessage(u[0], uids), conn)
		srv.Info("err ", err)
	}
}
