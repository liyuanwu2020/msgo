package engine

type HandlerFunc func(ctx *Context)

type MiddlewareFunc func(handlerFunc HandlerFunc) HandlerFunc
