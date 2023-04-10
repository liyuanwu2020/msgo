package engine

import (
	"github.com/liyuanwu2020/msgo/mslog"
	"log"
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
	//for _, group := range e.routerGroups {
	//	//去掉uri的分组名称
	//	routerName := SubStringLast(ctx.R.URL.Path, "/"+group.name)
	//	//路由是否存在
	//	node := group.treeNode.Get(routerName)
	//	if node != nil && node.isEnd {
	//		handlerFunc, ok := group.handlerFuncMap[node.routerName]
	//		//mslog.Printf("handlerFuncMap [%s] match [%s] %v", routerName, node.routerName, ok)
	//		if ok {
	//			ctx.NodeRouterName = node.routerName
	//			if handle, ok := handlerFunc[method]; ok {
	//				ctx.RequestMethod = method
	//				group.methodHandler(handle, ctx)
	//				return
	//			}
	//
	//			if handle, ok := handlerFunc[ANY]; ok {
	//				ctx.RequestMethod = ANY
	//				group.methodHandler(handle, ctx)
	//				return
	//			}
	//			ctx.W.WriteHeader(http.StatusMethodNotAllowed)
	//			log.Printf("%s %s not allowed", ctx.R.RequestURI, method)
	//			return
	//		}
	//	}
	//}
	ctx.W.WriteHeader(http.StatusNotFound)
	log.Println(method)
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
	engine := &Engine{
		router: router{},
	}
	return engine
}
