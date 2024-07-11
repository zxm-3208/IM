package group

import (
	"IM/apps/im/rpc/imclient"
	"IM/apps/social/api/internal/svc"
	"IM/apps/social/api/internal/types"
	"IM/apps/social/rpc/socialclient"
	"IM/pkg/constants"
	"IM/pkg/ctxdata"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupPutInHandleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupPutInHandleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupPutInHandleLogic {
	return &GroupPutInHandleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupPutInHandleLogic) GroupPutInHandle(req *types.GroupPutInHandleRep) (resp *types.GroupPutInHandleResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	res, err := l.svcCtx.Social.GroupPutInHandle(l.ctx, &socialclient.GroupPutInHandleReq{
		GroupReqId:   req.GroupReqId,
		GroupId:      req.GroupId,
		HandleUid:    uid,
		HandleResult: req.HandleResult,
	})

	if constants.HandlerResult(req.HandleResult) != constants.PassHandlerResult {
		return
	}

	if err != nil {
		return nil, err
	}

	// TODO: 通过后的业务：如发送通知等

	_, err = l.svcCtx.ImRpc.SetUpUserConversation(l.ctx, &imclient.SetUpUserConversationReq{
		SendId:   res.UserId,
		RecvId:   req.GroupId,
		ChatType: int32(constants.GroupChatType),
	})

	return nil, err
}
