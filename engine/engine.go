package engine

import (
	"github.com/liyuanwu2020/msgo/mslog"
	"net/http"
)

type Engine struct {
	router
	Logger       *mslog.Logger
	middlewares  []MiddlewareFunc
	errorHandler ErrorHandler
}

type ErrorHandler func(err error) (int, any)

func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := &Context{
		W: writer,
		R: request,
	}
	method := ctx.R.Method
	requestPath := ctx.R.URL.Path
	if node := e.router.node.Get(requestPath); node != nil {
		handlerFuncMap, ok := e.handlerFuncMap[node.routerName]
		if ok {
			ctx.NodeRouterName = node.routerName
			for _, v := range []string{method, ANY} {
				if handle, ok := handlerFuncMap[v]; ok {
					ctx.RequestMethod = v
					handle(ctx)
					return
				}
			}
			ctx.W.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
	ctx.W.WriteHeader(http.StatusNotFound)
}

// Run TLS use []string{certFile, keyFile}
func (e *Engine) Run(addr string, file ...string) {
	var err error
	if len(file) == 2 {
		err = http.ListenAndServeTLS(addr, file[0], file[1], e)
	} else {
		http.Handle("/", e)
		err = http.ListenAndServe(addr, nil)
	}

	if err != nil {
		e.Logger.Error(err)
	}
}

func Default() *Engine {
	engine := New()
	engine.Logger = mslog.Default()

	return engine
}

func New() *Engine {
	r := router{}
	r.handlerFuncMap = make(map[string]map[string]HandlerFunc, 10)
	engine := &Engine{
		router: r,
	}
	return engine
}
