package main

import (
	"IM/pkg/configserver"
	"IM/pkg/resultx"
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"sync"

	"IM/apps/user/api/internal/config"
	"IM/apps/user/api/internal/handler"
	"IM/apps/user/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/dev/user.yaml", "the config file")

var wg sync.WaitGroup

func main() {
	flag.Parse()

	var c config.Config
	//conf.MustLoad(*configFile, &c)
	err := configserver.NewConfigServer(*configFile, configserver.NewSail(&configserver.Config{
		ETCDEndpoints: "139.9.214.194:3379",
		ProjectKey:    "98c6f2c2287f4c73cea3d40ae7ec3ff2",
		Namespace:     "user",
		Configs:       "user-api.yaml",
		//ConfigFilePath: "./etc/conf",
		LogLevel: "DEBUG",
	})).MustLoad(&c, func(bytes []byte) error {
		var c config.Config
		configserver.LoadFromJsonBytes(bytes, &c)

		proc.WrapUp() // 停止服务的接收，并重新启动服务

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
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	httpx.SetErrorHandlerCtx(resultx.ErrHandler(c.Name))
	httpx.SetOkHandler(resultx.OkHandler)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
