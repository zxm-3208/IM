package handler

import (
	"IM/apps/im/ws/internal/handler/user"
	"IM/apps/im/ws/internal/svc"
	"IM/apps/im/ws/websocket"
)

func RegisterHandlers(srv *websocket.Server, svc *svc.ServiceContext) {
	srv.AddRoute([]websocket.Route{
		{
			Method:  "user.online",
			Handler: user.OnLine(svc),
		},
	})
}
