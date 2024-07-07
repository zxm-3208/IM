package svc

import (
	"IM/apps/user/models"
	"IM/apps/user/rpc/internal/config"
	"IM/pkg/constants"
	"IM/pkg/ctxdata"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"time"
)

type ServiceContext struct {
	Config config.Config
	*redis.Redis
	UserModels models.UsersModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewMysql(c.Mysql.DataSource)

	return &ServiceContext{
		Config:     c,
		Redis:      redis.MustNewRedis(c.Redisx),
		UserModels: models.NewUsersModel(sqlConn, c.Cache),
	}
}

func (svc *ServiceContext) SetRootToken() error {
	// 生成jwt todo: jwt过期处理
	systemToken, err := ctxdata.GetJwtToken(svc.Config.Jwt.AccessSecret, time.Now().Unix(), 3600*24*365, constants.SYSTEM_ROOT_UID)
	if err != nil {
		return err
	}
	// 写入到redis
	return svc.Redis.Set(constants.REDIS_SYSTEM_ROOT_TOEKN, systemToken)
}
