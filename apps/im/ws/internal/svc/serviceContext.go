package svc

import (
	"IM/apps/im/immodels"
	"IM/apps/im/ws/internal/config"
	"IM/apps/task/mq/mqclient"
)

type ServiceContext struct {
	Config                config.Config
	ChatLogModel          immodels.ChatLogModel
	MsgChatTransferClient mqclient.MsgChatTransferClient
}

func NewServiceContext(config config.Config) *ServiceContext {
	return &ServiceContext{
		Config:                config,
		ChatLogModel:          immodels.MustChatLogModel(config.Mongo.Url, config.Mongo.Db),
		MsgChatTransferClient: mqclient.NewMsgChatTransferClient(config.MsgChatTransfer.Addrs, config.MsgChatTransfer.Topic),
	}
}
