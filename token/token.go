package token

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/liyuanwu2020/msgo"
	"time"
)

type JwtHandler struct {
	Alg string
	//过期时间
	Timeout        time.Duration
	RefreshTimeout time.Duration
	//时间函数
	TimeFun func() time.Time
	//秘钥
	Key []byte
	//私钥
	privateKey string
	//cookie存储
	SendCookie     bool
	CookieName     string
	CookieMaxAge   int
	CookieDomain   string
	SecureCookie   bool
	CookieHTTPOnly bool
	//登录认证
	Authenticator func(ctx *msgo.Context) (map[string]any, error)
}

type JwtResponse struct {
	Token        string
	RefreshToken string
}

func (j *JwtHandler) LogoutHandler(ctx *msgo.Context) error {
	//清除cookie即可
	if j.SendCookie {
		if j.CookieName == "" {
			j.CookieName = "jwt_token" //JWTToken
		}
		ctx.SetCookie(j.CookieName, "", -1, "/", j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
		return nil
	}
	return nil
}

func (j *JwtHandler) LoginHandler(ctx *msgo.Context) (*JwtResponse, error) {
	data, err := j.Authenticator(ctx)
	if err != nil {
		return nil, err
	}
	if j.Alg == "" {
		j.Alg = "HS256"
	}
	if j.TimeFun == nil {
		j.TimeFun = func() time.Time {
			return time.Now()
		}
	}
	signingMethod := jwt.GetSigningMethod(j.Alg)
	token := jwt.New(signingMethod)
	claims := token.Claims.(jwt.MapClaims)
	if data != nil {
		for k, v := range data {
			claims[k] = v
		}
	}
	expire := j.TimeFun().Add(j.Timeout)
	//
	claims["exp"] = expire.Unix()
	claims["iat"] = j.TimeFun().Unix()
	var tokenStr string
	var tokenErr error
	if j.usingPublicKeyAlgo() {
		tokenStr, tokenErr = token.SignedString(j.privateKey)
	} else {
		tokenStr, tokenErr = token.SignedString(j.Key)
	}
	if tokenErr != nil {
		return nil, tokenErr
	}
	refreshTokenStr, tokenErr := j.refreshToken(token)
	if tokenErr != nil {
		return nil, tokenErr
	}
	req := &JwtResponse{
		Token:        tokenStr,
		RefreshToken: refreshTokenStr,
	}
	if j.SendCookie {
		if j.CookieName == "" {
			j.CookieName = "jwt_token"
		}
		if j.CookieMaxAge == 0 {
			j.CookieMaxAge = int(expire.Unix() - j.TimeFun().Unix())
		}
		maxAge := j.CookieMaxAge
		ctx.SetCookie(j.CookieName, tokenStr, maxAge, "/", j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
	}
	return req, nil

}

func (j *JwtHandler) usingPublicKeyAlgo() bool {
	switch j.Alg {
	case "RS256", "RS512", "RS384":
		return true
	}
	return false
}

func (j *JwtHandler) refreshToken(token *jwt.Token) (string, error) {
	claims := token.Claims.(jwt.MapClaims)
	expire := j.TimeFun().Add(j.Timeout)
	claims["exp"] = expire.Unix()
	var tokenStr string
	var tokenErr error
	if j.usingPublicKeyAlgo() {
		tokenStr, tokenErr = token.SignedString(j.privateKey)
	} else {
		tokenStr, tokenErr = token.SignedString(j.Key)
	}
	if tokenErr != nil {
		return "", tokenErr
	}
	return tokenStr, nil
}
