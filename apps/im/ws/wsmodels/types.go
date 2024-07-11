package wsmodels

import (
	"IM/pkg/constants"
)

type (
	Msg struct {
		MsgId           string                 `mapstructure:"msgId"`
		constants.MType `mapstructure:"mType"` // 特别注意，mapstructure中间不能有空格
		ReadRecords     map[string]string      `mapstructure:"readRecords"`
		Content         string                 `mapstructure:"content"`
	}

	Chat struct {
		ConversationId string             `mapstructure:"conversationId"`
		SendId         string             `mapstructure:"sendId"`
		RecvId         string             `mapstructure:"recvId"`
		ChatType       constants.ChatType `mapstructure:"chatType"`
		SendTime       int64              `mapstructure:"sendTime"`
		Msg            `mapstructure:"msg"`
	}

	Push struct {
		// 消息类型: 1. 私聊、 2. 群聊
		ChatType constants.ChatType `mapstructure:"chatType"`
		// 会话id
		ConversationId string `mapstructure:"conversationId"`
		// 发送者
		SendId string `mapstructure:"sendId"`
		// 接收者
		RecvId  string   `mapstructure:"recvId"`
		RecvIds []string `mapstructure:"recvIds"`
		// 消息类型
		constants.MType `mapstructure:"mType"`
		// 内容
		Content string `mapstructure:"content"`
		// 发送时间
		SendTime int64 `mapstructure:"sendTime"`
		// 已读记录
		ReadRecords map[string]string `mapstructure:"readRecords"`
		// 消息ID
		MsgId string `mapstructure:"msgId"`
		// 内容类型(正常的消息, 已读/未读消息)
		ContentType constants.ContentType `mapstructure:"contentType"`
	}

	// 用于接收客户端已读的消息
	MarkRead struct {
		constants.ChatType `mapstructure:"chatType"`
		RecvId             string   `mapstructure:"recvId"`
		ConversationId     string   `mapstructure:"conversationId"`
		MsgIds             []string `mapstructure:"msgIds"`
	}
)
