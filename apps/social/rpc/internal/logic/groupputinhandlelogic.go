package logic

import (
	"IM/apps/social/socialmodels"
	"IM/pkg/constants"
	"IM/pkg/xerr"
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"IM/apps/social/rpc/internal/svc"
	"IM/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	ErrGroupReqBeforePass   = xerr.NewMsgErr("请求已通过")
	ErrGroupReqBeforeRefuse = xerr.NewMsgErr("请求已拒绝")
)

type GroupPutInHandleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupPutInHandleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupPutInHandleLogic {
	return &GroupPutInHandleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupPutInHandleLogic) GroupPutInHandle(in *social.GroupPutInHandleReq) (*social.GroupPutInHandleResp, error) {
	groupReq, err := l.svcCtx.GroupRequestsModel.FindOne(l.ctx, uint64(in.GroupReqId))
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find friend req err %v req %v", err, in.GroupReqId)
	}

	switch constants.HandlerResult(groupReq.HandleResult.Int64) {
	case constants.PassHandlerResult:
		return nil, errors.WithStack(ErrGroupReqBeforePass)
	case constants.RefuseHandlerResult:
		return nil, errors.WithStack(ErrGroupReqBeforeRefuse)
	}

	groupReq.HandleResult = sql.NullInt64{
		Int64: int64(in.HandleResult),
		Valid: true,
	}

	err = l.svcCtx.GroupRequestsModel.Trans(
		l.ctx, func(ctx context.Context, session sqlx.Session) error {
			if err := l.svcCtx.GroupRequestsModel.Update(l.ctx, session, groupReq); err != nil {
				return errors.Wrapf(xerr.NewDBErr(), "update friend req err %v req %v", err, groupReq)
			}
			if constants.HandlerResult(groupReq.HandleResult.Int64) != constants.PassHandlerResult {
				return nil
			}

			var operatorUid sql.NullString
			if in.HandleUid != "" {
				operatorUid.String = in.HandleUid
				operatorUid.Valid = true
			}

			groupMember := &socialmodels.GroupMembers{
				GroupId:     groupReq.GroupId,
				UserId:      groupReq.ReqId,
				RoleLevel:   int64(constants.AtLargeGroupRoleLevel),
				OperatorUid: operatorUid,
			}

			_, err = l.svcCtx.GroupMembersModel.Insert(l.ctx, session, groupMember)
			if err != nil {
				return errors.Wrapf(xerr.NewDBErr(), "insert friend err %v req %v", err, groupMember)
			}

			return nil
		})

	return &social.GroupPutInHandleResp{}, nil
}
