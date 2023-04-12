package engine

import (
	"fmt"
	"github.com/liyuanwu2020/msgo/engine/gateway"
	"github.com/liyuanwu2020/msgo/mslog"
	"net/http"
	"net/http/httputil"
	"net/url"
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
		W:      writer,
		R:      request,
		Logger: e.Logger,
	}
	method := ctx.R.Method
	requestPath := ctx.R.URL.Path

	if node := e.node.Get(requestPath); node != nil {
		//网关的处理逻辑
		if e.gatewayConfigs != nil {
			gwConfig, _ := e.gatewayConfigMap[node.routerName]
			target, _ := url.Parse(fmt.Sprintf("http://%s:%d%s", gwConfig.Host, gwConfig.Port, requestPath))
			e.Logger.Info(target)
			return

			director := func(request *http.Request) {

			}
			response := func(response *http.Response) error {

				return nil
			}
			handler := func(writer http.ResponseWriter, request *http.Request, err error) {

			}
			proxy := httputil.ReverseProxy{Director: director, ModifyResponse: response, ErrorHandler: handler}
			proxy.ServeHTTP(writer, request)
			return
		}

		handlerFuncMap, ok := e.handlerFuncMap[node.routerName]
		if ok {
			ctx.NodeRouterName = node.routerName
			for _, v := range []string{method, ANY} {
				if handle, ok := handlerFuncMap[v]; ok {
					ctx.RequestMethod = v
					e.methodHandler(handle, ctx)
					return
				}
			}
			ctx.W.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
	ctx.W.WriteHeader(http.StatusNotFound)
}

func (e *Engine) methodHandler(h HandlerFunc, ctx *Context) {
	//包裹引擎级别中间件
	for _, middleware := range e.middlewares {
		h = middleware(h)
	}

	//包裹方法级别中间件
	funcMiddles := e.router.middlewaresFuncMap[ctx.NodeRouterName][ctx.RequestMethod]
	if funcMiddles != nil {
		funcLen := len(funcMiddles) - 1
		for i := funcLen; i > -1; i-- {
			middleware := funcMiddles[i]
			h = middleware(h)
		}
	}
	//执行方法
	h(ctx)
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

func (e *Engine) Use(middlewareFunc ...MiddlewareFunc) {
	e.middlewares = append(e.middlewares, middlewareFunc...)
}

func (e *Engine) SetGateConfigs(configs []gateway.GWConfig) {
	e.gatewayConfigs = configs
	if len(configs) > 0 {
		for _, config := range configs {
			e.node.Put(config.Path, config.Name)
			e.gatewayConfigMap[config.Name] = config
		}
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
	r.middlewaresFuncMap = make(map[string]map[string][]MiddlewareFunc, 10)
	r.gatewayConfigMap = make(map[string]gateway.GWConfig, 10)
	engine := &Engine{
		router: r,
	}
	return engine
}
