package ctxdata

import "github.com/golang-jwt/jwt"

const Identify = "uid"

// stringKey: 密钥; iat: 当前时间; seconds: 过期时间; uid: 用户id
func GetJwtToken(secretKey string, iat, seconds int64, uid string) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims[Identify] = uid

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims

	return token.SignedString([]byte(secretKey))
}
