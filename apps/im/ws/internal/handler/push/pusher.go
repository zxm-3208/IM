package push

import (
	"IM/apps/im/ws/internal/svc"
	"IM/apps/im/ws/websocket"
	"IM/apps/im/ws/wsmodels"
	"IM/pkg/constants"
	"github.com/mitchellh/mapstructure"
)

/*
WSserver将消息推送给目标服务(客户端)
*/

func Push(svcCtx *svc.ServiceContext) websocket.HandlerFunc {
	return func(srv *websocket.Server, conn *websocket.Conn, msg *websocket.Message) {
		var data wsmodels.Push
		if err := mapstructure.Decode(msg.Data, &data); err != nil {
			srv.Send(websocket.NewErrorMessage(err))
			return
		}
		// 发送的目标
		switch data.ChatType {
		case constants.SingleChatType:
			single(srv, &data, data.RecvId)
		case constants.GroupChatType:
			group(srv, &data)
		}
	}
}

func single(srv *websocket.Server, data *wsmodels.Push, recvId string) error {
	// 发送的目标
	rconn := srv.GetConn(recvId)
	if rconn == nil {
		return nil
	}

	srv.Infof("push msg %v", data)

	return srv.Send(websocket.NewMessage(data.SendId, &wsmodels.Chat{
		ConversationId: data.ConversationId,
		ChatType:       data.ChatType,
		SendTime:       data.SendTime,
		Msg: wsmodels.Msg{
			ReadRecords: data.ReadRecords,
			MsgId:       data.MsgId,
			MType:       data.MType,
			Content:     data.Content,
		},
	}), rconn)
}

// 并行发送
func group(srv *websocket.Server, data *wsmodels.Push) error {
	for _, id := range data.RecvIds {
		func(id string) {
			srv.Schedule(func() { // TaskRunner接口的方法
				single(srv, data, id)
			})
		}(id)
	}
	return nil
}
