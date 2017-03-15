package govan

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Params map[string][]string

func (v Params) Get(key string) string {
	if v == nil {
		return ""
	}
	vs := v[key]
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

func (v Params) Set(key, value string) {
	v[key] = []string{value}
}

func (v Params) Add(key, value string) {
	v[key] = append(v[key], value)
}

func (v Params) Del(key string) {
	delete(v, key)
}

type Router interface {
	Routes() Handler
	Use(method, path string, handler ...Handler)
	GET(path string, handler ...Handler)
	POST(path string, handler ...Handler)
	PUT(path string, handler ...Handler)
	DELETE(path string, handler ...Handler)
	HEAD(path string, handler ...Handler)
}

func NewRouter() Router {
	return &router{
		router: httprouter.New(),
	}
}

type routerError struct {
	error
}

type router struct {
	router *httprouter.Router
}

func (r *router) Use(method, path string, handler ...Handler) {
	r.router.Handle(method, path, func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
		if c, ok := r.Context().Value(govanCtxKey).(*Ctx); ok {
			p := Params{}
			for _, param := range params {
				p.Add(param.Key, param.Value)
			}
			c.Map(p)

			c.fork(handler...)
		}
	})
}

func (r *router) GET(path string, handler ...Handler) {
	r.Use("GET", path, handler...)
}

func (r *router) POST(path string, handler ...Handler) {
	r.Use("POST", path, handler...)
}

func (r *router) PUT(path string, handler ...Handler) {
	r.Use("PUT", path, handler...)
}

func (r *router) DELETE(path string, handler ...Handler) {
	r.Use("DELETE", path, handler...)
}

func (r *router) HEAD(path string, handler ...Handler) {
	r.Use("HEAD", path, handler...)
}

func (r *router) Routes() Handler {
	return func(cx *Ctx) error {
		if handle, params, tsr := r.router.Lookup(cx.Req.Method, cx.Req.URL.Path); handle != nil {
			handle(cx.Res, cx.Req, params)
			return cx.err
		} else if tsr && cx.Req.URL.Path != "/" {
			path := cx.Req.URL.Path
			if len(path) > 1 && path[len(path)-1] == '/' {
				cx.Req.URL.Path = path[:len(path)-1]
			} else {
				cx.Req.URL.Path = path + "/"
			}
			cx.Header("Location", cx.Req.URL.String())
			cx.Status = 301
		}

		return cx.Next()
	}
}
