package msgTransfer

import (
	"IM/apps/im/immodels"
	"IM/apps/task/mq/internal/svc"
	"IM/pkg/constants"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"sync"
	"time"
)

type msgMergeInsert struct {
	svcCtx         *svc.ServiceContext
	mu             sync.Mutex
	conversationId string
	chatLogs       []*immodels.ChatLog
	chatLogCh      chan *immodels.ChatLog
	//count          int
	// 上次推送时间
	pushTime time.Time
	done     chan struct{}
}

func newMsgMergeInsert(chatLog *immodels.ChatLog, chatLogCh chan *immodels.ChatLog, svc *svc.ServiceContext) *msgMergeInsert {
	m := &msgMergeInsert{
		conversationId: chatLog.ConversationId,
		chatLogs:       make([]*immodels.ChatLog, 0, GroupMsgInsertRecordDelayCount+10),
		chatLogCh:      chatLogCh,
		//count:          1,
		pushTime: time.Now(),
		done:     make(chan struct{}),
		svcCtx:   svc,
	}
	m.chatLogs = append(m.chatLogs, chatLog)

	go m.transfer()
	return m
}

// 合并消息
func (m *msgMergeInsert) mergeChatLog(chatLog *immodels.ChatLog) {
	m.mu.Lock()
	defer m.mu.Unlock()
	//m.count++
	m.chatLogs = append(m.chatLogs, chatLog)
}

func (m *msgMergeInsert) transfer() {
	timer := time.NewTimer(GroupMsgInsertRecordDelayTime / 2)
	defer timer.Stop()
	var ctx = context.Background()
	for {
		select {
		case <-m.done:
			return
		// 定时器检测
		case <-timer.C:
			m.mu.Lock()
			pushTime := m.pushTime
			val := GroupMsgInsertRecordDelayTime - time.Since(pushTime)
			chatLogs := m.chatLogs
			// 未到达释放要求
			if val > 0 || chatLogs == nil {
				if val > 0 {
					timer.Reset(val)
				}
				m.mu.Unlock()
				continue
			}
			// 达到释放要求
			m.pushTime = time.Now()
			m.chatLogs = nil
			//m.count = 0
			timer.Reset(GroupMsgInsertRecordDelayTime / 2)
			// 推送
			logx.Infof("超过等待时间合并条件，写入数据库  %v", chatLogs)
			m.mu.Unlock()
			m.svcCtx.ChatLogModel.InsertMany(ctx, chatLogs)
			m.svcCtx.ConversationModel.UpdateMsg(ctx, chatLogs[len(chatLogs)-1])
		default:
			//m.mu.Lock()
			//if m.count >= GroupMsgInsertRecordDelayCount {
			//	chatLogs := m.chatLogs
			//	m.chatLogs = nil
			//	m.count = 0
			//
			//	logx.Infof("达到合并量, 写入数据库 %v", chatLogs)
			//	m.mu.Unlock()
			//	m.svcCtx.ChatLogModel.InsertMany(ctx, chatLogs)
			//	m.svcCtx.ConversationModel.UpdateMsg(ctx, chatLogs[len(chatLogs)-1])
			//	continue
			//}

			// 该对象长时间没有达到释放要求，清空并推送消息以节省资源
			if m.IsIdle() {
				//m.mu.Unlock()
				//使得MsgInsertTransfer 清理
				m.chatLogCh <- &immodels.ChatLog{
					ChatType:       constants.GroupChatType,
					ConversationId: m.conversationId,
				}
				continue
			}
			//m.mu.Unlock()
			//tempDelay := GroupMsgInsertRecordDelayTime / 4
			//if tempDelay > time.Second {
			//	tempDelay = time.Second
			//}
			//time.Sleep(tempDelay)
		}
	}
}

func (m *msgMergeInsert) IsIdle() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isIdle()
}

func (m *msgMergeInsert) isIdle() bool {
	pushTime := m.pushTime
	val := GroupMsgInsertRecordDelayTime*2 - time.Since(pushTime)
	if val <= 0 && m.chatLogs == nil {
		//if val <= 0 && m.chatLogs == nil && m.count == 0 {
		return true
	}
	return false
}

func (m *msgMergeInsert) clear() {
	select {
	case <-m.done:
	default:
		close(m.done)
	}
	m.chatLogs = nil
}
