package govan

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type Render interface {
	JSON(interface{}) error
	HTML(name string, data interface{}) error
	Text(data string) error
	Redirect(url string, status ...int)
}

type render struct {
	tpl *template.Template
}

func (r *render) json(c *Ctx, i interface{}) error {
	b, err := json.Marshal(i)

	if err != nil {
		return err
	}

	if c.Req.Header.Get("Content-Type") == "" {
		c.Type("application/json")
	}

	c.Body(b)

	return nil
}

func (r *render) html(c *Ctx, name string, binding interface{}) error {
	buf := new(bytes.Buffer)

	if err := r.tpl.ExecuteTemplate(buf, name, binding); err != nil {
		return err
	}

	if c.Req.Header.Get("Content-Type") == "" {
		c.Type("text/html")
	}

	c.Body(buf.Bytes())

	return nil
}

func (r *render) text(c *Ctx, data string) error {
	c.Type("text/plain").Body([]byte(data))
	return nil
}

func (r *render) redirect(c *Ctx, url string, status ...int) {
	if len(status) == 0 && c.Status == 0 {
		c.Status = 302
	} else if len(status) > 0 {
		c.Status = status[0]
	}

	http.Redirect(c.Res, c.Req, url, c.Status)
}

type RenderOptions struct {
	Directory string
	Funcs     []template.FuncMap
}

type ctxRender struct {
	*render
	ctx *Ctx
}

func (r *ctxRender) JSON(i interface{}) error {
	return r.json(r.ctx, i)
}

func (r *ctxRender) HTML(name string, i interface{}) error {
	return r.html(r.ctx, name, i)
}

func (r *ctxRender) Text(data string) error {
	return r.text(r.ctx, data)
}

func (r *ctxRender) Redirect(url string, status ...int) {
	r.redirect(r.ctx, url, status...)
}

func NewRenderProvider(opts ...RenderOptions) Handler {
	ren := &render{}

	if len(opts) > 0 {
		opt := opts[0]

		if opt.Directory != "" {
			tpl := template.New(opt.Directory)
			template.Must(tpl.Parse("Govan"))

			filepath.Walk(opt.Directory, func(path string, info os.FileInfo, err error) error {
				r, err := filepath.Rel(opt.Directory, path)
				if err != nil {
					panic(err)
				}
				ext := filepath.Ext(r)
				if ext == ".html" {
					buf, err := ioutil.ReadFile(path)

					if err != nil {
						panic(err)
					}

					name := (r[0 : len(r)-len(ext)])
					tmpl := tpl.New(filepath.ToSlash(name))

					for _, funcs := range opt.Funcs {
						tmpl.Funcs(funcs)
					}

					template.Must(tmpl.Parse(string(buf)))
				}
				return nil
			})

			ren.tpl = tpl
		}
	}

	return func(c *Ctx) {
		var r Render = &ctxRender{ren, c}
		c.MapTo(r, (*Render)(nil))
		c.Next()
	}
}
