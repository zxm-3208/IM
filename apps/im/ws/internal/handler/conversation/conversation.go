package conversation

import (
	"IM/apps/im/ws/internal/svc"
	"IM/apps/im/ws/websocket"
	"IM/apps/im/ws/wsmodels"
	"IM/apps/task/mq/mq"
	"IM/pkg/constants"
	"github.com/mitchellh/mapstructure"
	"time"
)

// 针对用户处理的消息
func Chat(srvCtx *svc.ServiceContext) websocket.HandlerFunc {
	return func(srv *websocket.Server, conn *websocket.Conn, msg *websocket.Message) {
		var data wsmodels.Chat

		if err := mapstructure.Decode(msg.Data, &data); err != nil {
			srv.Send(websocket.NewErrorMessage(err), conn)
			return
		}

		switch data.ChatType {
		case constants.SingleChatType:
			// 消息由wsServer发送给MQclient
			err := srvCtx.MsgChatTransferClient.Push(&mq.MsgChatTransfer{
				ConversationId: data.ConversationId,
				ChatType:       data.ChatType,
				SendId:         conn.Uid,
				RecvId:         data.RecvId,
				SendTime:       time.Now().UnixNano(),
				MsgType:        data.Msg.MType,
				MsgContent:     data.Msg.Content,
			})
			if err != nil {
				srv.Send(websocket.NewErrorMessage(err), conn)
				return
			}
		}

		//l := logic.NewConversation(context.Background(), srv, srvCtx)
		//if err := l.SingleChat(&data, srv.GetUsers(conn)[0]); err != nil {
		//	srv.Send(websocket.NewErrorMessage(err), conn)
		//}

		//srv.SendByUserId(websocket.NewMessage(conn.Uid, wsmodels.Chat{
		//	ConversationId: data.ConversationId,
		//	SendId:         data.SendId,
		//	RecvId:         data.RecvId,
		//	ChatType:       data.ChatType,
		//	SendTime:       time.Now().UnixNano(),
		//	Msg:            data.Msg,
		//}), data.RecvId)
	}
}
