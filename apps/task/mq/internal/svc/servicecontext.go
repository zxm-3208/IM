package svc

import (
	"IM/apps/im/immodels"
	"IM/apps/im/ws/websocket"
	"IM/apps/task/mq/internal/config"
	"IM/pkg/constants"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"net/http"
)

type ServiceContext struct {
	Config config.Config

	WsClient websocket.Client
	*redis.Redis

	immodels.ChatLogModel
}

func NewServiceContext(config config.Config) *ServiceContext {
	svc := &ServiceContext{
		Config:       config,
		Redis:        redis.MustNewRedis(config.Redisx),
		ChatLogModel: immodels.MustChatLogModel(config.Mongo.Url, config.Mongo.Db),
	}

	token, err := svc.GetSystemToken()
	if err != nil {
		panic(err)
	}

	header := http.Header{}
	header.Set("Authorization", token)
	svc.WsClient = websocket.NewClient(config.Ws.Host, websocket.WithClientHeader(header))
	return svc
}

func (svc *ServiceContext) GetSystemToken() (string, error) {
	return svc.Redis.Get(constants.REDIS_SYSTEM_ROOT_TOEKN)
}
