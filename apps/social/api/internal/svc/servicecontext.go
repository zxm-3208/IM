package svc

import (
	"IM/apps/im/rpc/imclient"
	"IM/apps/social/api/internal/config"
	"IM/apps/social/rpc/socialclient"
	"IM/apps/user/rpc/userclient"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	UserRpc userclient.User
	Social  socialclient.Social
	ImRpc   imclient.Im
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,

		UserRpc: userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),

		Social: socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),

		ImRpc: imclient.NewIm(zrpc.MustNewClient(c.ImRpc)),
	}
}
