package msgTransfer

import (
	"IM/apps/im/immodels"
	"IM/apps/im/ws/websocket"
	"IM/apps/task/mq/internal/svc"
	"IM/apps/task/mq/mq"
	"IM/pkg/constants"
	"context"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
)

// kafka消费者
type MsgChatTransfer struct {
	logx.Logger
	svcCtx *svc.ServiceContext
}

func NewMsgChatTransfer(svc *svc.ServiceContext) *MsgChatTransfer {
	return &MsgChatTransfer{
		Logger: logx.WithContext(context.Background()),
		svcCtx: svc,
	}
}

// 只要类型的方法签名与接口中定义的方法签名完全匹配，那么该类型就自动实现了接口，无需在类型定义中显式声明它实现了哪个接口“鸭子类型” (kq.queue文件中的接口)
func (m *MsgChatTransfer) Consume(key, value string) error {
	fmt.Println("key:", key, "value:", value)

	var (
		data mq.MsgChatTransfer
		ctx  = context.Background()
	)

	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return err
	}

	// 记录数据
	if err := m.addChatLog(ctx, &data); err != nil {
		return err
	}

	// 推送消息(推送到mq的Wsclient中)
	return m.svcCtx.WsClient.Send(websocket.Message{
		Type:   websocket.FrameData,
		Method: "push",
		FromId: constants.SYSTEM_ROOT_UID,
		Data:   data,
	})

	return nil
}

func (m *MsgChatTransfer) addChatLog(ctx context.Context, data *mq.MsgChatTransfer) error {
	// 记录消息
	chatLog := immodels.ChatLog{
		ConversationId: data.ConversationId,
		SendId:         data.SendId,
		RecvId:         data.RecvId,
		ChatType:       data.ChatType,
		MsgType:        data.MsgType,
		MsgFrom:        0,
		MsgContent:     data.MsgContent,
		SendTime:       data.SendTime,
	}
	return m.svcCtx.ChatLogModel.Insert(ctx, &chatLog)
}
