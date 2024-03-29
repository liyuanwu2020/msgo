package msgo

import (
	"encoding/base64"
	"net/http"
)

type Accounts struct {
	UnAuthHandler func(ctx *Context)
	Users         map[string]string
}

func (a *Accounts) BasicAuth(next HandlerFunc) HandlerFunc {
	return func(ctx *Context) {
		username, password, ok := ctx.R.BasicAuth()
		ctx.Logger.Info([]string{username, password})
		if !ok {
			a.unAuthHandler(ctx)
			return
		}
		pwd, exists := a.Users[username]
		if !exists || pwd != password {
			a.unAuthHandler(ctx)
			return
		}

		ctx.Set("user", username)
		next(ctx)
	}
}

func (a *Accounts) unAuthHandler(ctx *Context) {
	if a.UnAuthHandler != nil {
		a.UnAuthHandler(ctx)
	} else {
		ctx.W.WriteHeader(http.StatusUnauthorized)
	}
}

func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
