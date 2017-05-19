package govan

import (
	"net/http"

	"github.com/unrolled/render"
)

type Render interface {
	JSON(interface{}) error
	HTML(name string, data interface{}) error
	Text(data string) error
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

func (r *ctxRender) JSON(i interface{}) error {
	return r.r.JSON(r.ctx.Res, 200, i)
}

func (r *ctxRender) HTML(name string, i interface{}) error {
	return r.r.HTML(r.ctx.Res, 200, name, i)
}

func (r *ctxRender) Text(data string) error {
	return r.r.Text(r.ctx.Res, 200, data)
}

func (r *ctxRender) Redirect(url string, status ...int) {
	s := 302

	if len(status) > 0 {
		s = status[0]
	}

	http.Redirect(r.ctx.Res, r.ctx.Req, url, s)
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
