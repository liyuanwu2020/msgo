package msgo

import (
	"log"
	"net/http"
)

const ANY = "ANY"

type HandlerFunc func(ctx *Context)

type MiddlewareFunc func(handlerFunc HandlerFunc) HandlerFunc

type routerGroup struct {
	name               string
	handlerFuncMap     map[string]map[string]HandlerFunc
	middlewaresFuncMap map[string]map[string][]MiddlewareFunc
	handlerMethodMap   map[string][]string
	treeNode           *treeNode
	middlewares        []MiddlewareFunc
}

func (r *routerGroup) Use(middlewareFunc ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, middlewareFunc...)
}

func (r *routerGroup) methodHandler(h HandlerFunc, ctx *Context) {
	//执行方法前置中间件
	if r.middlewares != nil {
		for _, middleware := range r.middlewares {
			h = middleware(h)
		}
	}
	//执行方法级别的中间件
	funcMiddles := r.middlewaresFuncMap[ctx.NodeRouterName][ctx.RequestMethod]
	if funcMiddles != nil {
		for _, middleware := range funcMiddles {
			h = middleware(h)
		}
	}
	//执行主业务程序
	h(ctx)
}

func (r *routerGroup) request(name string, handlerFunc HandlerFunc, method string, middlewareFunc ...MiddlewareFunc) {
	_, ok := r.handlerFuncMap[name]
	if !ok {
		r.handlerFuncMap[name] = make(map[string]HandlerFunc)
		r.middlewaresFuncMap[name] = make(map[string][]MiddlewareFunc)
	}
	_, ok = r.handlerFuncMap[name][method]
	if ok {
		panic(name + " 有重复的路由")
	}
	r.handlerFuncMap[name][method] = handlerFunc
	r.handlerMethodMap[method] = append(r.handlerMethodMap[method], name)
	r.middlewaresFuncMap[name][method] = append(r.middlewaresFuncMap[name][method], middlewareFunc...)
	r.treeNode.Put(name)
}

func (r *routerGroup) Any(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.request(name, handlerFunc, ANY, middlewareFunc...)
}

func (r *routerGroup) Get(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.request(name, handlerFunc, http.MethodGet, middlewareFunc...)
}

func (r *routerGroup) Post(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.request(name, handlerFunc, http.MethodPost, middlewareFunc...)
}

func (r *routerGroup) Put(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.request(name, handlerFunc, http.MethodPut, middlewareFunc...)
}

func (r *routerGroup) Delete(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.request(name, handlerFunc, http.MethodDelete, middlewareFunc...)
}

func (r *routerGroup) Patch(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.request(name, handlerFunc, http.MethodPatch, middlewareFunc...)
}
func (r *routerGroup) Head(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.request(name, handlerFunc, http.MethodHead, middlewareFunc...)
}
func (r *routerGroup) Options(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.request(name, handlerFunc, http.MethodOptions, middlewareFunc...)
}

type router struct {
	routerGroups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	g := &routerGroup{
		name:               name,
		handlerFuncMap:     make(map[string]map[string]HandlerFunc),
		middlewaresFuncMap: make(map[string]map[string][]MiddlewareFunc),
		handlerMethodMap:   make(map[string][]string),
		treeNode:           &treeNode{name: "/", children: make([]*treeNode, 0)},
	}
	r.routerGroups = append(r.routerGroups, g)
	return g
}

type Engine struct {
	router
}

func New() *Engine {
	return &Engine{
		router: router{},
	}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.httpRequestHandle(w, r)
}

func (e *Engine) Run() {

	http.Handle("/", e)
	err := http.ListenAndServe(":8088", nil)
	if err != nil {
		log.Fatal("启动失败", err)
	}
}

func (e *Engine) httpRequestHandle(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	ctx := &Context{W: w, R: r}
	for _, group := range e.routerGroups {
		//去掉uri的分组名称
		routerName := SubStringLast(r.RequestURI, "/"+group.name)
		//路由是否存在
		node := group.treeNode.Get(routerName)
		if node != nil && node.isEnd {
			handlerFunc, ok := group.handlerFuncMap[node.routerName]
			log.Printf("handlerFuncMap routerName %s %v", node.routerName, ok)
			if ok {
				ctx.NodeRouterName = node.routerName
				if handle, ok := handlerFunc[method]; ok {
					ctx.RequestMethod = method
					group.methodHandler(handle, ctx)
					return
				}

				if handle, ok := handlerFunc[ANY]; ok {
					ctx.RequestMethod = ANY
					group.methodHandler(handle, ctx)
					return
				}
				w.WriteHeader(http.StatusMethodNotAllowed)
				log.Printf("%s %s not allowed", r.RequestURI, method)
				return
			}
		}
	}
	w.WriteHeader(http.StatusNotFound)
	log.Printf("%s %s not found", r.RequestURI, method)
}
