package engine

import (
	"fmt"
	"github.com/liyuanwu2020/msgo/mslog"
	"net"
	"strings"
	"time"
)

type HandlerFunc func(ctx *Context)

type MiddlewareFunc func(handlerFunc HandlerFunc) HandlerFunc

func Logging(next HandlerFunc) HandlerFunc {
	cnf := mslog.LoggerConfig{
		Formatter: mslog.DefaultFormatter,
		Out:       mslog.DefaultWriter,
	}
	return func(ctx *Context) {
		ctx.Logger.Info("执行顺序 logging")

		// Start timer
		start := time.Now()
		path := ctx.R.URL.Path
		raw := ctx.R.URL.RawQuery
		//执行业务
		next(ctx)
		// stop timer
		stop := time.Now()
		latency := stop.Sub(start)
		ip, _, _ := net.SplitHostPort(strings.TrimSpace(ctx.R.RemoteAddr))
		clientIP := net.ParseIP(ip)
		if raw != "" {
			path = path + "?" + raw
		}
		param := &mslog.LogFormatterParams{
			Request:    ctx.R,
			TimeStamp:  stop,
			StatusCode: ctx.StatusCode,
			Latency:    latency,
			ClientIP:   clientIP,
			Method:     ctx.R.Method,
			Path:       path,
		}
		_, err := fmt.Fprint(cnf.Out, cnf.Formatter(param))
		if err != nil {
			return
		}
	}
}
