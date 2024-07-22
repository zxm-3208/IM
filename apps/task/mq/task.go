package main

import (
	"IM/apps/task/mq/internal/config"
	"IM/apps/task/mq/internal/handler"
	"IM/apps/task/mq/internal/svc"
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"net/http"
)

var configFile = flag.String("f", "etc/dev/task.yaml", "the config file")

func main() {
	flag.Parse()

	go func() {
		http.ListenAndServe("0.0.0.0:8002", nil)
	}()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	if err := c.SetUp(); err != nil {
		panic(err)
	}
	svcCtx := svc.NewServiceContext(c)
	listen := handler.NewListen(svcCtx)

	serviceGroup := service.NewServiceGroup()
	//defer serviceGroup.Stop()
	for _, s := range listen.Services() {
		serviceGroup.Add(s)
	}
	fmt.Println("Starting mqueue server at ...")
	serviceGroup.Start()
}
