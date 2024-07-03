package handler

import (
	"IM/apps/im/ws/internal/svc"
	"IM/pkg/ctxdata"
	"context"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/token"
	"net/http"
)

type JwtAuth struct {
	srvCtx *svc.ServiceContext
	parser *token.TokenParser
	logx.Logger
}

func NewJwtAuth(srvCtx *svc.ServiceContext) *JwtAuth {
	return &JwtAuth{
		srvCtx: srvCtx,
		parser: token.NewTokenParser(),
		Logger: logx.WithContext(context.Background()),
	}
}

func (j *JwtAuth) Auth(w http.ResponseWriter, r *http.Request) bool {
	tok, err := j.parser.ParseToken(r, j.srvCtx.Config.JwtAuth.AccessSecret, "")
	if err != nil {
		j.Errorf("parse token err %v ", err)
		return false
	}
	if !tok.Valid { // 检验token是否有效
		j.Errorf("parse valid err %v ", err)
		return false
	}

	j.Infof("token %v ", tok)

	j.Infof("token %v ", tok.Claims)

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		j.Errorf("parse claims err %v ", claims)
		return false
	}

	*r = *r.WithContext(context.WithValue(r.Context(), ctxdata.Identify, claims[ctxdata.Identify]))

	return true
}

func (j *JwtAuth) UserId(r *http.Request) string {
	return ctxdata.GetUId(r.Context())
}
