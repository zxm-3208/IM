package main

import (
	"IM/apps/im/ws/internal/config"
	"IM/apps/im/ws/internal/handler"
	"IM/apps/im/ws/internal/svc"
	"IM/apps/im/ws/websocket"
	"IM/pkg/configserver"
	"IM/pkg/constants"
	"IM/pkg/ctxdata"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
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
	pid := os.Getpid()
	fmt.Printf("进程 PID: %d \n", pid)

	go func() {
		http.ListenAndServe("0.0.0.0:8000", nil)
	}()

	if err := c.SetUp(); err != nil {
		panic(err)
	}

	ctx := svc.NewServiceContext(c)

	// 设置服务认证的token
	token, err := ctxdata.GetJwtToken(c.JwtAuth.AccessSecret, time.Now().Unix(), 3153600000, fmt.Sprintf("ws:%s", time.Now().Unix()))
	if err != nil {
		panic(err)
	}

	opts := []websocket.Options{
		websocket.WithAuthentication(handler.NewJwtAuth(ctx)),
		websocket.WithMaxConnectionIdle(10 * time.Minute),
		websocket.WithServerAck(websocket.OnlyAck),
		websocket.WithServerDiscover(websocket.NewRedisDiscover(http.Header{
			"Authorization": []string{token},
		}, constants.REDIS_DISCOVER_SRV, c.Redisx)),
	}

	srv := websocket.NewServer(c.ListenOn, opts...)
	defer srv.Stop()

	handler.RegisterHandlers(srv, ctx)

	fmt.Printf("Starting websocket server at %v ...\n", c.ListenOn)
	srv.Start()
}
