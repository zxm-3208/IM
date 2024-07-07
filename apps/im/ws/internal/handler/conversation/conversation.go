package conversation

import (
	"IM/apps/im/ws/internal/logic"
	"IM/apps/im/ws/internal/svc"
	"IM/apps/im/ws/websocket"
	"IM/apps/im/ws/wsmodels"
	"context"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"time"
)

// 针对用户处理的消息
func Chat(srvCtx *svc.ServiceContext) websocket.HandlerFunc {
	return func(srv *websocket.Server, conn *websocket.Conn, msg *websocket.Message) {
		var data wsmodels.Chat

		if err := mapstructure.Decode(msg.Data, &data); err != nil {
			fmt.Println("11111111111111111")
			srv.Send(websocket.NewErrorMessage(err), conn)
			return
		}

		l := logic.NewConversation(context.Background(), srv, srvCtx)
		if err := l.SingleChat(&data, srv.GetUsers(conn)[0]); err != nil {
			fmt.Println("222222222222222222")
			srv.Send(websocket.NewErrorMessage(err), conn)
		}

		srv.SendByUserId(websocket.NewMessage(conn.Uid, wsmodels.Chat{
			ConversationId: data.ConversationId,
			SendId:         data.SendId,
			RecvId:         data.RecvId,
			ChatType:       data.ChatType,
			SendTime:       time.Now().UnixNano(),
			Msg:            data.Msg,
		}), data.RecvId)
	}
}
