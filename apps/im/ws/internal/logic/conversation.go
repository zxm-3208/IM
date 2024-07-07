package logic

import (
	"IM/apps/im/immodels"
	"IM/apps/im/ws/internal/svc"
	"IM/apps/im/ws/websocket"
	"IM/apps/im/ws/wsmodels"
	"IM/pkg/wuid"
	"context"
	"time"
)

type Conversation struct {
	ctx    context.Context
	srv    *websocket.Server
	svcCtx *svc.ServiceContext
}

func NewConversation(ctx context.Context, srv *websocket.Server, svcCtx *svc.ServiceContext) *Conversation {
	return &Conversation{
		ctx:    ctx,
		srv:    srv,
		svcCtx: svcCtx,
	}
}

func (l *Conversation) SingleChat(data *wsmodels.Chat, userId string) error {
	if data.ConversationId == "" {
		data.ConversationId = wuid.ConbineId(userId, data.RecvId)
	}

	//time.Sleep(time.Minute)
	// 记录消息
	ChatLog := immodels.ChatLog{
		ConversationId: data.ConversationId,
		SendId:         userId,
		RecvId:         data.RecvId,
		ChatType:       data.ChatType,
		MsgType:        data.MType,
		MsgFrom:        0,
		MsgContent:     data.Content,
		SendTime:       time.Now().UnixNano(),
	}
	err := l.svcCtx.ChatLogModel.Insert(l.ctx, &ChatLog)
	return err
}
