package logic

import (
	"IM/apps/user/rpc/internal/svc"
	"IM/apps/user/rpc/user"
	"IM/pkg/xerr"
	"context"
	"database/sql"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
)

var ErruserNotFound = xerr.New(xerr.SERVER_COMMON_ERROR, "user not found")

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *user.GetUserInfoReq) (*user.GetUserInfoResp, error) {

	userEntity, err := l.svcCtx.UserModels.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErruserNotFound
		}
		return nil, err
	}

	var resp user.UserEntity
	copier.Copy(&resp, userEntity)

	return &user.GetUserInfoResp{
		User: &resp,
	}, nil
}
