package wsmodels

import (
	"IM/pkg/constants"
)

type (
	Msg struct {
		constants.MType `mapstructure:"mType"` // 特别注意，mapstructure中间不能有空格
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
		ChatType constants.ChatType `json:"chatType"`
		// 会话id
		ConversationId string `json:"conversationId"`
		// 发送者
		SendId string `json:"sendId"`
		// 接收者
		RecvId string `json:"recvId"`

		constants.MType `mapstructure:"mType"`
		Content         string `mapstructure:"content"`
		// 发送时间
		SendTime int64 `json:"sendTime"`
	}
)
