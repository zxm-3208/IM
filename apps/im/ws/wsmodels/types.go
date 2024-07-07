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
)
