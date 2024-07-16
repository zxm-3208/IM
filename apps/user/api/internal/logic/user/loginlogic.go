package user

import (
	"IM/apps/user/rpc/user"
	"IM/pkg/constants"
	"context"
	"github.com/jinzhu/copier"

	"IM/apps/user/api/internal/svc"
	"IM/apps/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户登录
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	loginResp, err := l.svcCtx.User.Login(l.ctx, &user.LoginReq{
		Phone:    req.Phone,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	var res types.LoginResp
	copier.Copy(&res, loginResp)
	// 设置用户在线
	l.svcCtx.Redis.HsetCtx(l.ctx, constants.REDIS_ONLINE_USER, loginResp.Id, "1")
	return &res, nil

}
