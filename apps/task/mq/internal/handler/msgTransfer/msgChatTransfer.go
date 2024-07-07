package msgTransfer

import (
	"IM/apps/task/mq/internal/svc"
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
)

// kafka消费者
type MsgChatTransfer struct {
	logx.Logger
	svcCtx *svc.ServiceContext
}

func NewMsgChatTransfer(svc *svc.ServiceContext) *MsgChatTransfer {
	return &MsgChatTransfer{
		Logger: logx.WithContext(context.Background()),
		svcCtx: svc,
	}
}

// 只要类型的方法签名与接口中定义的方法签名完全匹配，那么该类型就自动实现了接口，无需在类型定义中显式声明它实现了哪个接口“鸭子类型” (kq.queue文件中的接口)
func (m *MsgChatTransfer) Consume(key, value string) error {
	fmt.Println("key:", key, "value:", value)
	return nil
}
