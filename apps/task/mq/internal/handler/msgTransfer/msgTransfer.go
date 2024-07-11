package msgTransfer

import (
	"IM/apps/im/ws/websocket"
	"IM/apps/im/ws/wsmodels"
	"IM/apps/social/rpc/socialclient"
	"IM/apps/task/mq/internal/svc"
	"IM/pkg/constants"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type baseMsgTransfer struct {
	logx.Logger
	svcCtx *svc.ServiceContext
}

func NewBaseMsgTransfer(svc *svc.ServiceContext) *baseMsgTransfer {
	return &baseMsgTransfer{
		Logger: logx.WithContext(context.Background()),
		svcCtx: svc,
	}
}

func (m *baseMsgTransfer) Transfer(ctx context.Context, data *wsmodels.Push) error {
	switch data.ChatType {
	case constants.SingleChatType:
		return m.single(ctx, data)
	case constants.GroupChatType:
		return m.group(ctx, data)
	}
	return nil
}

func (m *baseMsgTransfer) single(ctx context.Context, data *wsmodels.Push) error {
	return m.svcCtx.WsClient.Send(websocket.Message{
		Type:   websocket.FrameData,
		Method: "push",
		FromId: constants.SYSTEM_ROOT_UID,
		Data:   data,
	})
}

func (m *baseMsgTransfer) group(ctx context.Context, data *wsmodels.Push) error {
	res, err := m.svcCtx.Social.GroupUsers(ctx, &socialclient.GroupUsersReq{
		GroupId: data.RecvId,
	})
	if err != nil {
		return err
	}

	data.RecvIds = make([]string, 0, len(res.List))
	for _, member := range res.List {
		// 跳过发送人
		if member.UserId == data.SendId {
			continue
		}
		data.RecvIds = append(data.RecvIds, member.UserId)
	}
	return m.svcCtx.WsClient.Send(websocket.Message{
		Type:   websocket.FrameData,
		Method: "push",
		FromId: constants.SYSTEM_ROOT_UID,
		Data:   data,
	})
}
