package govan

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/eknkc/govan/inject"
	"github.com/go-playground/form"
)

const govanCtxKey ctxKey = 1

type ctxKey int

type Handler interface{}

type Ctx struct {
	inject.Injector
	Res    ResponseWriter
	Req    *http.Request
	Status int
	body   []byte
	next   *middleware
	err    error
}

func (c *Ctx) Header(header, value string) *Ctx {
	c.Res.Header().Add(header, value)
	return c
}

func (c *Ctx) Body(body []byte) *Ctx {
	c.body = body
	return c
}

func (c *Ctx) Type(mime string) *Ctx {
	return c.Header("Content-Type", mime)
}

func (c *Ctx) Cookie(cookie *http.Cookie) *Ctx {
	http.SetCookie(c.Res, cookie)
	return c
}

func (c *Ctx) Bind(v interface{}) error {
	if c.Req.Header.Get("Content-Type") == "application/json" {
		decoder := json.NewDecoder(c.Req.Body)
		if err := decoder.Decode(v); err != nil {
			return err
		}
	} else {
		decoder := form.NewDecoder()
		if err := c.Req.ParseForm(); err != nil {
			return err
		}
		if err := decoder.Decode(v, c.Req.Form); err != nil {
			return err
		}
	}

	return nil
}

func (c *Ctx) Next() error {
	if c.next != nil {
		next := c.next
		c.next = next.next
		c.err = c.serve(next.handler)
		return c.err
	}

	return nil
}

func (c *Ctx) fork(handler ...Handler) error {
	if len(handler) < 1 {
		return nil
	}

	mwHead := &middleware{handler: handler[0]}
	mwTail := mwHead

	for _, h := range handler[1:] {
		mwTail.next = &middleware{handler: h}
		mwTail = mwTail.next
	}

	mwTail.next = c.next
	c.next = mwHead

	return c.Next()
}

func (c *Ctx) serve(h Handler) error {
	vals, err := c.Invoke(h)

	if err != nil {
		panic(err)
	}

	if len(vals) == 0 {
		return c.err
	}

	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	err = c.err

	for _, val := range vals {
		if c.Status == 0 && val.Kind() == reflect.Int {
			c.Status = int(val.Int())
		}

		if val.Kind() == reflect.Interface && val.Type().Implements(errorInterface) {
			if !val.IsNil() {
				err = val.Interface().(error)
			} else {
				err = nil
			}
		}
	}

	return err
}

type Govan struct {
	inject.Injector
	head *middleware
	log  *log.Logger
}

func New() *Govan {
	g := &Govan{
		Injector: inject.New(),
		head:     &middleware{handler: topMiddleware},
		log:      log.New(os.Stdout, "[govan] ", 0),
	}

	g.Map(g.log)

	return g
}

func (g *Govan) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	c := &Ctx{
		Injector: inject.New(),
		Res:      &responseWriter{ResponseWriter: rw},
		next:     g.head,
	}

	r = r.WithContext(context.WithValue(r.Context(), govanCtxKey, c))
	c.Req = r

	c.Map(c.Req)
	c.MapTo(c.Res, (*http.ResponseWriter)(nil))
	c.Map(c)
	c.SetParent(g)

	c.Next()

	if c.Res.Written() {
		return
	}

	if c.Status == 0 {
		if c.body != nil {
			c.Status = 200
		} else {
			c.Status = 404
		}
	}

	rw.WriteHeader(c.Status)

	if c.body != nil {
		rw.Write(c.body)
	}
}

func (n *Govan) Run(addr string) {
	n.log.Printf("listening on %s", addr)
	n.log.Fatal(http.ListenAndServe(addr, n))
}

func (n *Govan) Use(handler ...Handler) {
	last := n.head

	for {
		if last.next != nil {
			last = last.next
		} else {
			break
		}
	}

	for _, h := range handler {
		last.next = &middleware{handler: h}
		last = last.next
	}
}

type middleware struct {
	handler Handler
	next    *middleware
}

func topMiddleware(c *Ctx) {
	if err := c.Next(); err != nil {
		if c.body == nil {
			c.body = []byte(err.Error())
		}

		if c.Status == 0 {
			c.Status = 500
		}
	}
}
