package handler

import (
	"IM/apps/task/mq/internal/handler/msgTransfer"
	"IM/apps/task/mq/internal/svc"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/service"
)

// 考虑消息队列机制还可能会存在其他的业务可能，所以提供Services方法输出service.Service数组类型以便于增加其他异步任务的处理
type Listen struct {
	svc *svc.ServiceContext
}

func NewListen(svc *svc.ServiceContext) *Listen {
	return &Listen{svc: svc}
}

func (l *Listen) Services() []service.Service {
	// 在构建服务框架或微服务架构时，其中每个服务可能需要与其他服务协作或共享某些配置。通过返回一个切片，可以方便地添加更多的服务实例到这个切片中，而不需要修改方法的返回类型或结构。
	return []service.Service{
		// 可以添加多个消费者
		kq.MustNewQueue(l.svc.Config.MsgChatTransfer, msgTransfer.NewMsgChatTransfer(l.svc)),
	}
}
