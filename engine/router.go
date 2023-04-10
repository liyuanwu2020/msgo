package engine

import "net/http"

type router struct {
	name             string
	handlerFuncMap   map[string]map[string]HandlerFunc
	handlerMethodMap map[string][]string
	node             treeNode
}

const ANY = "ANY"

func (r *router) Any(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.request(name, handlerFunc, ANY, middlewareFunc...)
}

func (r *router) Get(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.request(name, handlerFunc, http.MethodGet, middlewareFunc...)
}

func (r *router) request(name string, handlerFunc HandlerFunc, method string, middlewareFunc ...MiddlewareFunc) {
	_, ok := r.handlerFuncMap[name]
	if !ok {
		r.handlerFuncMap[name] = make(map[string]HandlerFunc)
	}
	_, ok = r.handlerFuncMap[name][method]
	if ok {
		panic(name + " 有重复的路由")
	}
	r.handlerFuncMap[name][method] = handlerFunc
	r.handlerMethodMap[method] = append(r.handlerMethodMap[method], name)
	r.node.Put(name)
}
