package conversation

import (
	"IM/apps/im/ws/internal/svc"
	"IM/apps/im/ws/websocket"
	"IM/apps/im/ws/wsmodels"
	"IM/apps/task/mq/mq"
	"IM/pkg/constants"
	"IM/pkg/wuid"
	"fmt"
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

		if data.ConversationId == "" {
			switch data.ChatType {
			case constants.SingleChatType:
				data.ConversationId = wuid.ConbineId(conn.Uid, data.RecvId)
			case constants.GroupChatType:
				data.ConversationId = data.RecvId
			}
		}

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
}

// 针对已读未读消息的处理
func MarkRead(svc *svc.ServiceContext) websocket.HandlerFunc {
	return func(srv *websocket.Server, conn *websocket.Conn, msg *websocket.Message) {
		var data wsmodels.MarkRead
		if err := mapstructure.Decode(msg.Data, &data); err != nil {
			srv.Send(websocket.NewErrorMessage(err), conn)
			return
		}
		// 将消息发送给已读未读消息队列
		err := svc.MsgReadTransferClient.Push(&mq.MsgMarkRead{
			ChatType:       data.ChatType,
			ConversationId: data.ConversationId,
			SendId:         conn.Uid,
			RecvId:         data.RecvId,
			MsgIds:         data.MsgIds,
		})
		fmt.Println(err)
		if err != nil {
			srv.Send(websocket.NewErrorMessage(err), conn)
			return
		}
	}
}
