package config

import "github.com/zeromicro/go-zero/core/service"

type Config struct {
	service.ServiceConf // 该服务本身不是go-zero定义的，但后续需要使用go-zero中的提供的功能，因此需要引入配置并调用其内部的初始化对配置加载

	ListenOn string

	JwtAuth struct {
		AccessSecret string
	}

	Mongo struct {
		Url string
		Db  string
	}

	MsgChatTransfer struct {
		Topic string
		Addrs []string
	}

	MsgReadTransfer struct {
		Topic string
		Addrs []string
	}
}
