package main

import (
	"IM/apps/im/ws/internal/config"
	"IM/apps/im/ws/internal/handler"
	"IM/apps/im/ws/internal/svc"
	"IM/apps/im/ws/websocket"
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/core/conf"
	"time"
)

var configFile = flag.String("f", "etc/dev/im.yaml", "config file")

func main() {
	flag.Parse() // 用于解析命令行参数

	var c config.Config
	conf.MustLoad(*configFile, &c)

	if err := c.SetUp(); err != nil {
		panic(err)
	}

	ctx := svc.NewServiceContext(c)

	srv := websocket.NewServer(c.ListenOn,
		websocket.WithAuthentication(handler.NewJwtAuth(ctx)),
		websocket.WithMaxConnectionIdle(10*time.Minute),
	)
	defer srv.Stop()

	handler.RegisterHandlers(srv, ctx)

	fmt.Printf("Starting websocket server at %v ...\n", c.ListenOn)
	srv.Start()
}
