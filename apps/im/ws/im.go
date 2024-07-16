package main

import (
	"IM/apps/im/ws/internal/config"
	"IM/apps/im/ws/internal/handler"
	"IM/apps/im/ws/internal/svc"
	"IM/apps/im/ws/websocket"
	"IM/pkg/configserver"
	"flag"
	"fmt"
	"sync"
	"time"
)

var configFile = flag.String("f", "etc/dev/im.yaml", "config file")
var wg sync.WaitGroup

func main() {
	flag.Parse() // 用于解析命令行参数

	var c config.Config
	//conf.MustLoad(*configFile, &c)
	err := configserver.NewConfigServer(*configFile, configserver.NewSail(&configserver.Config{
		ETCDEndpoints: "139.9.214.194:3379",
		ProjectKey:    "98c6f2c2287f4c73cea3d40ae7ec3ff2",
		Namespace:     "im",
		Configs:       "im-ws.yaml",
		//ConfigFilePath: "./etc/conf",
		LogLevel: "DEBUG",
	})).MustLoad(&c, func(bytes []byte) error {
		var c config.Config
		configserver.LoadFromJsonBytes(bytes, &c)
		wg.Add(1)
		go func(c config.Config) {
			defer wg.Done()
			Run(c)
		}(c)
		return nil
	})
	if err != nil {
		panic(err)
	}

	wg.Add(1)
	go func(c config.Config) {
		defer wg.Done()

		Run(c)
	}(c)

	wg.Wait()
}

func Run(c config.Config) {
	if err := c.SetUp(); err != nil {
		panic(err)
	}

	ctx := svc.NewServiceContext(c)

	srv := websocket.NewServer(c.ListenOn,
		websocket.WithAuthentication(handler.NewJwtAuth(ctx)),
		websocket.WithMaxConnectionIdle(10*time.Minute),
		websocket.WithServerAck(websocket.OnlyAck),
	)
	defer srv.Stop()

	handler.RegisterHandlers(srv, ctx)

	fmt.Printf("Starting websocket server at %v ...\n", c.ListenOn)
	srv.Start()
}
