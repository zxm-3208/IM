package logic

import (
	"IM/apps/im/api/internal/svc"
	"IM/apps/im/api/internal/types"
	"IM/apps/im/rpc/im"
	"IM/apps/social/rpc/socialclient"
	"IM/apps/user/rpc/user"
	"IM/pkg/bitmap"
	"IM/pkg/constants"
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatLogReadRecordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetChatLogReadRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatLogReadRecordsLogic {
	return &GetChatLogReadRecordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetChatLogReadRecordsLogic) GetChatLogReadRecords(req *types.GetChatLogReadRecordsReq) (resp *types.GetChatLogReadRecordsResp, err error) {
	chatlogs, err := l.svcCtx.Im.GetChatLog(l.ctx, &im.GetChatLogReq{
		MsgId: req.MsgId,
	})

	if err != nil || len(chatlogs.List) == 0 {
		return nil, err
	}

	var (
		chatLog = chatlogs.List[0]
		reads   = []string{chatLog.SendId}
		unreads []string
		ids     []string
	)
	fmt.Println(chatLog)
	fmt.Println(constants.ChatType(chatLog.ChatType))
	// 分别设置已读未读
	switch constants.ChatType(chatLog.ChatType) {
	case constants.SingleChatType:
		if len(chatLog.ReadRecords) == 0 || chatLog.ReadRecords[0] == 0 {
			unreads = []string{chatLog.RecvId}
		} else {
			reads = append(reads, chatLog.RecvId)
		}
		ids = []string{chatLog.RecvId, chatLog.SendId}
	case constants.GroupChatType:
		groupUsers, err := l.svcCtx.Social.GroupUsers(l.ctx, &socialclient.GroupUsersReq{
			GroupId: chatLog.RecvId,
		})
		if err != nil {
			return nil, err
		}

		bitmaps := bitmap.Load(chatLog.ReadRecords)
		for _, members := range groupUsers.List {
			ids = append(ids, members.UserId)

			if members.UserId == chatLog.SendId {
				continue
			}

			if bitmaps.IsSet(members.UserId) {
				reads = append(reads, members.UserId)
			} else {
				unreads = append(unreads, members.UserId)
			}
		}
	}

	userEntitys, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{
		Ids: ids,
	})
	if err != nil {
		return nil, err
	}
	userEntitysSet := make(map[string]*user.UserEntity, len(userEntitys.User))
	for i, entity := range userEntitys.User {
		userEntitysSet[entity.Id] = userEntitys.User[i]
	}

	// 设置手机号码
	for i, read := range reads {
		if u := userEntitysSet[read]; u != nil {
			reads[i] = u.Phone
		}
	}

	for i, unread := range unreads {
		if u := userEntitysSet[unread]; u != nil {
			unreads[i] = u.Phone
		}
	}

	return &types.GetChatLogReadRecordsResp{
		Reads:   reads,
		UnReads: unreads,
	}, nil
}
