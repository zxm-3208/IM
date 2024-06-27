package logic

import (
	"IM/apps/social/socialmodels"
	"IM/pkg/constants"
	"IM/pkg/xerr"
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"time"

	"IM/apps/social/rpc/internal/svc"
	"IM/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupPutinLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupPutinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupPutinLogic {
	return &GroupPutinLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupPutinLogic) GroupPutin(in *social.GroupPutinReq) (*social.GroupPutinResp, error) {
	// 存在三种进群方式
	// 1. 普通用户申请： 如果群无验证直接进入
	// 2. 群成员邀请: 如果群无验证直接进入
	// 3. 群管理员/群创建者邀请：直接进入群

	var (
		inviteGroupMember *socialmodels.GroupMembers // 邀请者
		userGroupMember   *socialmodels.GroupMembers // 被邀请者
		groupInfo         *socialmodels.Groups

		err error
	)

	// 查询被邀请者是否已经是群成员
	userGroupMember, err = l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.ReqId, in.GroupId)
	if err != nil && err != socialmodels.ErrNotFound {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find group member by groud id and req id err %v, req %v, %v", err, in.GroupId, in.ReqId)
	}
	if userGroupMember != nil {
		return &social.GroupPutinResp{}, nil
	}

	// 查询群申请是否已经存在
	groupReq, err := l.svcCtx.GroupRequestsModel.FindByGroupIdAndReqId(l.ctx, in.GroupId, in.ReqId)
	if err != nil && err != socialmodels.ErrNotFound {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find group req by groud id and user id err %v, req %v, %v", err,
			in.GroupId, in.ReqId)
	}
	if groupReq != nil {
		return &social.GroupPutinResp{}, nil
	}

	// 群申请Req
	groupReq = &socialmodels.GroupRequests{
		Id:      0,
		ReqId:   in.ReqId,
		GroupId: in.GroupId,
		ReqMsg: sql.NullString{
			String: in.ReqMsg,
			Valid:  true,
		},
		ReqTime: sql.NullTime{
			Time:  time.Unix(in.ReqTime, 0),
			Valid: true,
		},
		JoinSource: sql.NullInt64{
			Int64: int64(in.JoinSource),
			Valid: true,
		},
		InviterUserId: sql.NullString{
			String: in.InviterUid,
			Valid:  true,
		},
		HandleUserId: sql.NullString{
			String: in.InviterUid,
			Valid:  true,
		},
		HandleResult: sql.NullInt64{
			Int64: int64(constants.NoHandlerResult),
			Valid: true,
		},
	}

	// 创建群用户的函数
	createGroupMember := func() {
		if err != nil {
			return
		}
		err = l.createGroupMember(in)
	}

	// 获取组群信息
	groupInfo, err = l.svcCtx.GroupsModel.FindOne(l.ctx, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find group by group id err %v, req %v", err, in.GroupId)
	}

	// 判断申请入群是否需要验证

	// 不需要
	if !groupInfo.IsVerify {
		defer createGroupMember()

		groupReq.HandleResult = sql.NullInt64{
			Int64: int64(constants.PassHandlerResult),
			Valid: true,
		}
		return l.createGroupReq(groupReq)
	}

	// 判断进群方式（管理员、创建者邀请无需验证）
	if constants.GroupJoinSource(in.JoinSource) == constants.PutInGroupJoinSource {
		// 用户主动申请入群，需要验证
		return l.createGroupReq(groupReq)
	}

	// 邀请入群，判断邀请者身份
	inviteGroupMember, err = l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.InviterUid, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find group member by groud id and user id err %v, req %v",
			in.InviterUid, in.GroupId)
	}

	// 判断邀请者是否为创建者或者管理员
	if constants.GroupRoleLevel(inviteGroupMember.RoleLevel) == constants.CreatorGroupRoleLevel || constants.
		GroupRoleLevel(inviteGroupMember.RoleLevel) == constants.ManagerGroupRoleLevel {
		defer createGroupMember()

		groupReq.HandleResult = sql.NullInt64{
			Int64: int64(constants.PassHandlerResult),
			Valid: true,
		}
		groupReq.HandleUserId = sql.NullString{
			String: in.InviterUid,
			Valid:  true,
		}
		return l.createGroupReq(groupReq)
	}
	return l.createGroupReq(groupReq)
}

func (l *GroupPutinLogic) createGroupMember(in *social.GroupPutinReq) error {
	var operatorUid sql.NullString
	if in.InviterUid != "" {
		operatorUid.String = in.InviterUid
		operatorUid.Valid = true
	}

	groupMember := &socialmodels.GroupMembers{
		GroupId:     in.GroupId,
		UserId:      in.ReqId,
		RoleLevel:   int64(constants.AtLargeGroupRoleLevel),
		OperatorUid: operatorUid,
	}
	_, err := l.svcCtx.GroupMembersModel.Insert(l.ctx, nil, groupMember)
	if err != nil {
		return errors.Wrapf(xerr.NewDBErr(), "insert friend err %v req %v", err, groupMember)
	}
	return nil
}

func (l *GroupPutinLogic) createGroupReq(groupReq *socialmodels.GroupRequests) (*social.GroupPutinResp, error) {

	_, err := l.svcCtx.GroupRequestsModel.Insert(l.ctx, groupReq)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "insert group req err %v req %v", err, groupReq)
	}

	return &social.GroupPutinResp{}, nil
}
