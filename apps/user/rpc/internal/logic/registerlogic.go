package logic

import (
	"IM/apps/user/models"
	"IM/pkg/ctxdata"
	"IM/pkg/encrypt"
	"IM/pkg/wuid"
	"context"
	"database/sql"
	"errors"
	"time"

	"IM/apps/user/rpc/internal/svc"
	"IM/apps/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	ErrPhoneIsRegistered = errors.New("手机号已经注册过了")
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *user.RegisterReq) (*user.RegisterResp, error) {
	// 1. 验证用户是否注册， 根据手机号码验证
	userEntity, err := l.svcCtx.UserModels.FindByPhone(l.ctx, in.Phone)
	if err != nil && err != models.ErrNotFound {
		return nil, err
	}
	if userEntity != nil {
		return nil, ErrPhoneIsRegistered
	}
	// 2. 定义用户数据 TODO: 需要补充部分数据
	userEntity = &models.Users{
		Id:       wuid.GetUid(l.svcCtx.Config.Mysql.DataSource),
		Avatar:   in.Avatar,
		Nickname: in.Nickname,
		Phone:    in.Phone,
		//Status:   sql.NullInt64{},
		Sex: sql.NullInt64{
			Int64: int64(in.Sex),
			Valid: true,
		},
		//CreatedAt: sql.NullTime{},
		//UpdatedAt: sql.NullTime{},
	}

	// 3. 对密码进行加密
	if len(in.Password) > 0 {
		genPassword, err := encrypt.GenPasswordHash([]byte(in.Password))
		if err != nil {
			return nil, err
		}
		userEntity.Password = sql.NullString{
			String: string(genPassword),
			Valid:  true,
		}
	}

	// 4. 将信息保存到数据库中
	_, err = l.svcCtx.UserModels.Insert(l.ctx, userEntity)
	if err != nil {
		return nil, err
	}

	// 5. 生成token
	now := time.Now().Unix()
	token, err := ctxdata.GetJwtToken(l.svcCtx.Config.Jwt.AccessSecret, now, l.svcCtx.Config.Jwt.AccessExpire, userEntity.Id)
	if err != nil {
		return nil, err
	}

	return &user.RegisterResp{
		Token:  token,
		Expire: now + l.svcCtx.Config.Jwt.AccessExpire,
	}, nil
}
