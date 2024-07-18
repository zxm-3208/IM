package svc

import (
	"IM/apps/im/immodels"
	"IM/apps/im/ws/websocket"
	"IM/apps/social/rpc/socialclient"
	"IM/apps/task/mq/internal/config"
	"IM/pkg/constants"
	"IM/pkg/interceptor"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"net/http"
)

var retryPolicy = `{
	"methodConfig" : [{
		"name": [{
			"service": "social.social"
		}],
		"waitForReady": true,
		"retryPolicy": {
			"maxAttempts": 5,
			"initialBackoff": "0.001s",
			"maxBackoff": "0.002s",
			"backoffMultiplier": 1.0,
			"retryableStatusCodes": ["UNKNOWN", "DEADLINE_EXCEEDED"]
		}
	}]
}`

type ServiceContext struct {
	Config config.Config

	WsClient websocket.Client
	*redis.Redis

	immodels.ChatLogModel
	immodels.ConversationModel

	socialclient.Social
}

func NewServiceContext(config config.Config) *ServiceContext {
	svc := &ServiceContext{
		Config:            config,
		Redis:             redis.MustNewRedis(config.Redisx),
		ChatLogModel:      immodels.MustChatLogModel(config.Mongo.Url, config.Mongo.Db),
		ConversationModel: immodels.MustConversationModel(config.Mongo.Url, config.Mongo.Db),
		Social: socialclient.NewSocial(zrpc.MustNewClient(config.SocialRpc,
			zrpc.WithDialOption(grpc.WithDefaultServiceConfig(retryPolicy)),
			zrpc.WithUnaryClientInterceptor(interceptor.DefaultIdempotentClient),
		)),
	}

	token, err := svc.GetSystemToken()
	if err != nil {
		panic(err)
	}

	header := http.Header{}
	header.Set("Authorization", token)
	svc.WsClient = websocket.NewClient(config.Ws.Host,
		websocket.WithClientHeader(header),
		websocket.WithClientDiscover(websocket.NewRedisDiscover(header, constants.REDIS_DISCOVER_SRV, config.Redisx)),
	)
	return svc
}

func (svc *ServiceContext) GetSystemToken() (string, error) {
	return svc.Redis.Get(constants.REDIS_SYSTEM_ROOT_TOEKN)
}
