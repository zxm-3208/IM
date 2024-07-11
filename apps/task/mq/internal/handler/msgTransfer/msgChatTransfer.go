package msgTransfer

import (
	"IM/apps/im/immodels"
	"IM/apps/im/ws/wsmodels"
	"IM/apps/task/mq/internal/svc"
	"IM/apps/task/mq/mq"
	"IM/pkg/bitmap"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// kafka消费者
type MsgChatTransfer struct {
	*baseMsgTransfer
}

func NewMsgChatTransfer(svc *svc.ServiceContext) *MsgChatTransfer {
	return &MsgChatTransfer{
		NewBaseMsgTransfer(svc),
	}
}

// 只要类型的方法签名与接口中定义的方法签名完全匹配，那么该类型就自动实现了接口，无需在类型定义中显式声明它实现了哪个接口“鸭子类型” (kq.queue文件中的接口)
func (m *MsgChatTransfer) Consume(key, value string) error {
	fmt.Println("key:", key, "value:", value)

	var (
		data  mq.MsgChatTransfer
		ctx   = context.Background()
		msgId = primitive.NewObjectID()
	)

	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return err
	}

	// 记录数据
	if err := m.addChatLog(ctx, msgId, &data); err != nil {
		return err
	}

	return m.Transfer(ctx, &wsmodels.Push{
		ChatType:       data.ChatType,
		ConversationId: data.ConversationId,
		SendId:         data.SendId,
		RecvId:         data.RecvId,
		RecvIds:        data.RecvIds,
		MType:          data.MsgType,
		Content:        data.MsgContent,
		SendTime:       data.SendTime,
		MsgId:          msgId.Hex(),
	})
}

func (m *MsgChatTransfer) addChatLog(ctx context.Context, msgId primitive.ObjectID, data *mq.MsgChatTransfer) error {
	// 记录消息
	chatLog := immodels.ChatLog{
		ID:             msgId,
		ConversationId: data.ConversationId,
		SendId:         data.SendId,
		RecvId:         data.RecvId,
		ChatType:       data.ChatType,
		MsgType:        data.MsgType,
		MsgFrom:        0,
		MsgContent:     data.MsgContent,
		SendTime:       data.SendTime,
	}

	readRecords := bitmap.NewBitmap(0)
	readRecords.Set(chatLog.SendId)
	chatLog.ReadRecords = readRecords.Export()

	err := m.svcCtx.ChatLogModel.Insert(ctx, &chatLog)
	if err != nil {
		return err
	}
	return m.svcCtx.ConversationModel.UpdateMsg(ctx, &chatLog)
}
