package logic

import (
	"IM/pkg/xerr"
	"context"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"

	"IM/apps/social/rpc/internal/svc"
	"IM/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupListLogic {
	return &GroupListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupListLogic) GroupList(in *social.GroupListReq) (*social.GroupListResp, error) {

	// 用户所在的组群
	userGroup, err := l.svcCtx.GroupMembersModel.ListByUserId(l.ctx, in.UserId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "list group member err %v req %v", err, in.UserId)
	}
	if len(userGroup) == 0 {
		return &social.GroupListResp{}, nil
	}

	ids := make([]string, 0, len(userGroup))
	for _, v := range userGroup {
		ids = append(ids, v.GroupId)
	}

	// 组群信息
	groups, err := l.svcCtx.GroupsModel.ListByGroupIds(l.ctx, ids)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "list group err %v req %v", err, ids)
	}

	var respList []*social.Groups
	copier.Copy(&respList, &groups)

	return &social.GroupListResp{
		List: respList,
	}, nil
}
