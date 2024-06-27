package group

import (
	"IM/apps/social/rpc/socialclient"
	"IM/pkg/ctxdata"
	"context"

	"IM/apps/social/api/internal/svc"
	"IM/apps/social/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupPutInLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupPutInLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupPutInLogic {
	return &GroupPutInLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupPutInLogic) GroupPutIn(req *types.GroupPutInRep) (resp *types.GroupPutInResp, err error) {
	uid := ctxdata.GetUId(l.ctx)

	_, err = l.svcCtx.Social.GroupPutin(l.ctx, &socialclient.GroupPutinReq{
		GroupId: req.GroupId,
		ReqId:   uid,
		ReqMsg:  req.ReqMsg,
		ReqTime: req.ReqTime,
		//InviterUid: req.InviterUid,
		JoinSource: int32(req.JoinSource),
	})
	return
}
