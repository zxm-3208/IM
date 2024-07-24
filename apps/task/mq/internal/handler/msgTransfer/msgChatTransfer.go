package msgTransfer

import (
	"IM/apps/im/immodels"
	"IM/apps/im/ws/wsmodels"
	"IM/apps/task/mq/internal/svc"
	"IM/apps/task/mq/mq"
	"IM/pkg/bitmap"
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"sync"
	"time"
)

const (
	GroupMsgInsertHandlerAtTransfer = iota
	GroupMsgInsertHandlerDelayTransfer
)

var (
	GroupMsgInsertRecordDelayCount = 10
	GroupMsgInsertRecordDelayTime  = time.Second
)

// kafka消费者
type MsgChatTransfer struct {
	*baseMsgTransfer
	mu        sync.Mutex
	chatLogCh chan *immodels.ChatLog
	groupMsgs map[string]*msgMergeInsert
}

func NewMsgChatTransfer(svc *svc.ServiceContext) *MsgChatTransfer {
	m := &MsgChatTransfer{
		baseMsgTransfer: NewBaseMsgTransfer(svc),
		groupMsgs:       make(map[string]*msgMergeInsert, 1),
		chatLogCh:       make(chan *immodels.ChatLog, 1),
	}

	if svc.Config.MsgInsertHandler.GroupMsgInsertHandler != GroupMsgInsertHandlerAtTransfer {
		if svc.Config.MsgInsertHandler.GroupMsgInsertRecordDelayCount > 0 {
			GroupMsgInsertRecordDelayCount = svc.Config.MsgInsertHandler.GroupMsgInsertRecordDelayCount
		}

		if svc.Config.MsgInsertHandler.GroupMsgInsertRecordDelayTime > 0 {
			GroupMsgInsertRecordDelayTime = time.Duration(svc.Config.MsgInsertHandler.GroupMsgInsertRecordDelayTime) * time.Second
		}
	}
	if m.svcCtx.Config.MsgInsertHandler.GroupMsgInsertHandler != GroupMsgInsertHandlerAtTransfer {
		go m.transfer()
	}
	return m
}

// 只要类型的方法签名与接口中定义的方法签名完全匹配，那么该类型就自动实现了接口，无需在类型定义中显式声明它实现了哪个接口“鸭子类型” (kq.queue文件中的接口)
func (m *MsgChatTransfer) Consume(key, value string) error {
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
	chatLog := &immodels.ChatLog{
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
	// 没有开启消息合并，直接写入数据库
	if m.svcCtx.Config.MsgInsertHandler.GroupMsgInsertHandler == GroupMsgInsertHandlerAtTransfer {
		err := m.svcCtx.ChatLogModel.Insert(ctx, chatLog)
		if err != nil {
			return err
		}
		return m.svcCtx.ConversationModel.UpdateMsg(ctx, chatLog)
	}
	// 合并后写入数据库
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.groupMsgs[chatLog.ConversationId]; ok {
		// 存在该映射
		//m.Infof("merge chatLog %v", chatLog.ConversationId)
		// 合并请求
		m.groupMsgs[chatLog.ConversationId].mergeChatLog(chatLog)
	} else {
		m.Infof("newGroupMsgInsert push %v", chatLog.ConversationId)
		m.groupMsgs[chatLog.ConversationId] = newMsgMergeInsert(chatLog, m.chatLogCh, m.svcCtx)
	}

	return nil
}

func (m *MsgChatTransfer) transfer() {
	for chatLog := range m.chatLogCh {
		if _, ok := m.groupMsgs[chatLog.ConversationId]; ok && m.groupMsgs[chatLog.ConversationId].IsIdle() {
			m.groupMsgs[chatLog.ConversationId].clear()
			delete(m.groupMsgs, chatLog.ConversationId)
		}
	}
}
