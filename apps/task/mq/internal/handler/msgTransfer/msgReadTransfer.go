package msgTransfer

import (
	"IM/apps/im/ws/wsmodels"
	"IM/apps/task/mq/internal/svc"
	"IM/apps/task/mq/mq"
	"IM/pkg/bitmap"
	"IM/pkg/constants"
	"context"
	"encoding/base64"
	"encoding/json"
)

type MsgReadTransfer struct {
	*baseMsgTransfer
}

func NewMsgReadTransfer(svc *svc.ServiceContext) *MsgReadTransfer {
	return &MsgReadTransfer{
		baseMsgTransfer: NewBaseMsgTransfer(svc),
	}
}

func (m *MsgReadTransfer) Consume(key, value string) error {
	m.Info("MsgReadTransfer ", value)

	var (
		data mq.MsgMarkRead
		ctx  = context.Background()
	)
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return err
	}

	// 更新消息聊天记录中的已读记录
	readRecords, err := m.UpdateChatLogRead(ctx, &data)
	if err != nil {
		return err
	}
	return m.Transfer(ctx, &wsmodels.Push{
		ConversationId: data.ConversationId,
		ChatType:       data.ChatType,
		SendId:         data.SendId,
		RecvId:         data.RecvId,
		ContentType:    constants.ContentMakeRead,
		ReadRecords:    readRecords,
	})

}

func (m *MsgReadTransfer) UpdateChatLogRead(ctx context.Context, data *mq.MsgMarkRead) (map[string]string, error) {
	res := make(map[string]string)
	chatLogs, err := m.svcCtx.ChatLogModel.ListByMsgIds(ctx, data.MsgIds)
	if err != nil {
		return nil, err
	}
	m.Infof("chatLogs %v", chatLogs)

	// 处理已读
	for _, chatLog := range chatLogs {
		switch chatLog.ChatType {
		case constants.SingleChatType:
			chatLog.ReadRecords = []byte{1}
		case constants.GroupChatType:
			readRecords := bitmap.Load(chatLog.ReadRecords)
			readRecords.Set(data.SendId)
			chatLog.ReadRecords = readRecords.Export()
		}
		res[chatLog.ID.Hex()] = base64.StdEncoding.EncodeToString(chatLog.ReadRecords)

		err = m.svcCtx.ChatLogModel.UpdateMakeRead(ctx, chatLog.ID, chatLog.ReadRecords)
		if err != nil {
			m.Errorf("update make read err %v", err)
		}
	}
	return res, nil
}
