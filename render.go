package govan

import (
	"net/http"

	"github.com/unrolled/render"
)

type Render interface {
	JSON(data interface{}, status ...int) error
	JSONTo(rw http.ResponseWriter, data interface{}, status ...int) error
	HTML(name string, data interface{}, status ...int) error
	HTMLTo(rw http.ResponseWriter, name string, data interface{}, status ...int) error
	Text(data string, status ...int) error
	TextTo(rw http.ResponseWriter, data string, status ...int) error
	Redirect(url string, status ...int)
}

type renderer struct {
	r *render.Render
}

type RenderOptions struct {
	render.Options
}

type ctxRender struct {
	*renderer
	ctx *Ctx
}

func (r *ctxRender) JSON(i interface{}, status ...int) error {
	return r.r.JSON(r.ctx.Res, getStatus(200, status), i)
}

func (r *ctxRender) JSONTo(rw http.ResponseWriter, i interface{}, status ...int) error {
	return r.r.JSON(rw, getStatus(200, status), i)
}

func (r *ctxRender) HTML(name string, i interface{}, status ...int) error {
	return r.r.HTML(r.ctx.Res, getStatus(200, status), name, i)
}

func (r *ctxRender) HTMLTo(rw http.ResponseWriter, name string, i interface{}, status ...int) error {
	return r.r.HTML(rw, getStatus(200, status), name, i)
}

func (r *ctxRender) Text(data string, status ...int) error {
	return r.r.Text(r.ctx.Res, getStatus(200, status), data)
}

func (r *ctxRender) TextTo(rw http.ResponseWriter, data string, status ...int) error {
	return r.r.Text(rw, getStatus(200, status), data)
}

func (r *ctxRender) Redirect(url string, status ...int) {
	http.Redirect(r.ctx.Res, r.ctx.Req, url, getStatus(302, status))
}

func NewRenderProvider(opts ...RenderOptions) Handler {
	var ren *renderer

	if len(opts) > 0 {
		ren = &renderer{r: render.New(opts[0].Options)}
	} else {
		ren = &renderer{r: render.New()}
	}

	return func(c *Ctx) {
		var r Render = &ctxRender{ren, c}
		c.MapTo(r, (*Render)(nil))
		c.Next()
	}
}

func getStatus(def int, st []int) int {
	if len(st) > 0 {
		def = st[0]
	}
	return def
}
