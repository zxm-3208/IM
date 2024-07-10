package mq

import "IM/pkg/constants"

type MsgChatTransfer struct {
	// 消息类型:1. 私聊；2. 群聊
	ChatType constants.ChatType `json:"chatType"`
	// 会话id
	ConversationId string `json:"conversationId"`
	// 发送者
	SendId string `json:"sendId"`
	// 接收者
	RecvId string `json:"recvId"`

	RecvIds []string `json:"recvIds"`
	// 消息类型
	MsgType constants.MType `json:"msgTyp,omitempty"`
	// 消息内容
	MsgContent string `json:"msgContent,omitempty"`
	// 发送时间
	SendTime int64 `json:"sendTime"`
}
