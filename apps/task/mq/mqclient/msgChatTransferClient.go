package mqclient

import (
	"IM/apps/task/mq/mq"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-queue/kq"
)

/*
MQClient
MQ生产者,websocket将消息写入消息队列。
*/

type MsgChatTransferClient interface {
	Push(msg *mq.MsgChatTransfer) error
}

type msgChatTransferClient struct {
	pusher *kq.Pusher
}

func NewMsgChatTransferClient(addrs []string, topic string, opts ...kq.PushOption) *msgChatTransferClient {
	return &msgChatTransferClient{
		pusher: kq.NewPusher(addrs, topic),
	}
}

func (c *msgChatTransferClient) Push(msg *mq.MsgChatTransfer) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	fmt.Println("push", msg)
	return c.pusher.Push(string(body))
}
