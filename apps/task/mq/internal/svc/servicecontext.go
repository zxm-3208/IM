package svc

import "IM/apps/task/mq/internal/config"

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(config config.Config) *ServiceContext {
	return &ServiceContext{
		Config: config,
	}
}
