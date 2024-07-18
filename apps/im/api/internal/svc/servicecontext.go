package svc

import (
	"IM/apps/im/api/internal/config"
	"IM/apps/im/rpc/imclient"
	"IM/apps/social/rpc/socialclient"
	"IM/apps/user/rpc/userclient"
	"IM/pkg/interceptor"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

var SocialRetryPolicy = `{
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

var UserRetryPolicy = `{
	"methodConfig" : [{
		"name": [{
			"service": "user.User"
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

var ImRetryPolicy = `{
	"methodConfig" : [{
		"name": [{
			"service": "im.Im"
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

	imclient.Im
	userclient.User
	socialclient.Social
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		Im: imclient.NewIm(zrpc.MustNewClient(c.ImRpc,
			zrpc.WithDialOption(grpc.WithDefaultServiceConfig(ImRetryPolicy)),
			zrpc.WithUnaryClientInterceptor(interceptor.DefaultIdempotentClient),
		)),
		User: userclient.NewUser(zrpc.MustNewClient(c.UserRpc,
			zrpc.WithDialOption(grpc.WithDefaultServiceConfig(UserRetryPolicy)),
			zrpc.WithUnaryClientInterceptor(interceptor.DefaultIdempotentClient),
		)),
		Social: socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc,
			zrpc.WithDialOption(grpc.WithDefaultServiceConfig(SocialRetryPolicy)),
			zrpc.WithUnaryClientInterceptor(interceptor.DefaultIdempotentClient),
		)),
	}
}
