package msgo

import (
	"html/template"
	"net/http"
)

type Context struct {
	W              http.ResponseWriter
	R              *http.Request
	NodeRouterName string
	RequestMethod  string
	engine         *Engine
}

func (c *Context) HTML(status int, html string) error {
	c.W.Header().Set("Content-Type", "text/html;charset=utf-8")
	//状态码默认200 http.StatusOK
	c.W.WriteHeader(status)
	_, err := c.W.Write([]byte(html))
	return err
}

func (c *Context) HTMLTemplate(name string, data any, files ...string) error {
	c.W.Header().Set("Content-Type", "text/html;charset=utf-8")
	t := template.New(name)
	parseFiles, err := t.ParseFiles(files...)
	if err != nil {
		return err
	}
	err = parseFiles.Execute(c.W, data)
	return err
}

func (c *Context) HTMLTemplateGlob(name string, data any, pattern string) error {
	c.W.Header().Set("Content-Type", "text/html;charset=utf-8")
	t := template.New(name)
	parseFiles, err := t.ParseGlob(pattern)
	if err != nil {
		return err
	}
	err = parseFiles.Execute(c.W, data)
	return err
}

func (c *Context) Template(name string, data any) error {
	c.W.Header().Set("Content-Type", "text/html;charset=utf-8")
	var err error
	err = c.engine.HTMLRender.Template.ExecuteTemplate(c.W, name, data)
	return err
}
