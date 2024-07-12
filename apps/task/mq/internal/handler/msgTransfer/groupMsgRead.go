package msgTransfer

import (
	"IM/apps/im/ws/wsmodels"
	"IM/pkg/constants"
	"github.com/zeromicro/go-zero/core/logx"
	"sync"
	"time"
)

type groupMsgRead struct {
	mu             sync.Mutex
	conversationId string
	push           *wsmodels.Push
	pushCh         chan *wsmodels.Push
	count          int
	// 上次推送时间
	pushTime time.Time
	done     chan struct{}
}

func newGroupMsgRead(push *wsmodels.Push, pushCh chan *wsmodels.Push) *groupMsgRead {
	m := &groupMsgRead{
		conversationId: push.ConversationId,
		push:           push,
		pushCh:         pushCh,
		count:          1,
		pushTime:       time.Now(),
		done:           make(chan struct{}),
	}

	go m.transfer()
	return m
}

// 合并消息
func (m *groupMsgRead) mergePush(push *wsmodels.Push) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.count++
	for msgId, read := range push.ReadRecords {
		m.push.ReadRecords[msgId] = read
	}
}

func (m *groupMsgRead) transfer() {
	timer := time.NewTimer(GroupMsgReadRecordDelayTime / 2)
	defer timer.Stop()

	for {
		select {
		case <-m.done:
			return
		// 定时器检测
		case <-timer.C:
			m.mu.Lock()
			pushTime := m.pushTime
			val := GroupMsgReadRecordDelayTime - time.Since(pushTime)
			push := m.push
			logx.Infof("timer.C %v val %v", time.Now(), val)
			// 未到达释放要求
			if val > 0 && m.count < GroupMsgReadRecordDelayCount || push == nil {
				if val > 0 {
					timer.Reset(val)
				}
				m.mu.Unlock()
				continue
			}
			// 达到释放要求
			m.pushTime = time.Now()
			m.push = nil
			m.count = 0
			timer.Reset(GroupMsgReadRecordDelayTime / 2)
			m.mu.Unlock()

			// 推送
			logx.Infof("超过等待时间合并条件，进行推送 %v", push)
			m.pushCh <- push
		default:
			m.mu.Lock()
			if m.count >= GroupMsgReadRecordDelayCount {
				push := m.push
				m.push = nil
				m.count = 0
				m.mu.Unlock()

				// 推送
				logx.Infof("达到推送量, 进行推送 %v", push)
				m.pushCh <- push
				continue
			}

			// 该对象长时间没有达到释放要求，清空并推送消息以节省资源
			if m.isIdle() {
				m.mu.Unlock()
				// 使得MsgReadTransfer 清理
				m.pushCh <- &wsmodels.Push{
					ChatType:       constants.GroupChatType,
					ConversationId: m.conversationId,
				}
				continue
			}

			m.mu.Unlock()
			tempDelay := GroupMsgReadRecordDelayTime / 4
			if tempDelay > time.Second {
				tempDelay = time.Second
			}
			time.Sleep(tempDelay)
		}

	}
}

func (m *groupMsgRead) IsIdle() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isIdle()
}

func (m *groupMsgRead) isIdle() bool {
	pushTime := m.pushTime
	val := GroupMsgReadRecordDelayTime*2 - time.Since(pushTime)

	if val <= 0 && m.push == nil && m.count == 0 {
		return true
	}
	return false
}

func (m *groupMsgRead) clear() {
	select {
	case <-m.done:
	default:
		close(m.done)
	}
	m.push = nil
}
