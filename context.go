package msgo

import (
	"github.com/liyuanwu2020/msgo/render"
	"html/template"
	"net/http"
	"net/url"
)

type Context struct {
	W              http.ResponseWriter
	R              *http.Request
	NodeRouterName string
	RequestMethod  string
	engine         *Engine
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

// HTML 直接输出 html 字符串
func (c *Context) HTML(status int, html string) error {
	return c.Render(&render.HTML{Data: html}, status)
}

// Template 带模板的 html
func (c *Context) Template(name string, data any) error {
	return c.Render(&render.HTML{
		Name:       name,
		Data:       data,
		Template:   c.engine.HTMLRender.Template,
		IsTemplate: true,
	}, http.StatusOK)
}

func (c *Context) JSON(status int, data any) error {
	return c.Render(&render.JSON{Data: data}, status)
}

func (c *Context) XML(status int, data any) error {
	return c.Render(&render.XML{Data: data}, status)
}

// String 字符串
func (c *Context) String(status int, format string, values ...any) error {
	c.W.WriteHeader(status)
	err := c.Render(&render.String{
		Format: format,
		Values: values,
	}, status)
	return err
}

// Redirect 重定向
func (c *Context) Redirect(status int, location string) error {
	return c.Render(&render.Redirect{
		StatusCode: status,
		Request:    c.R,
		Location:   location,
	}, status)
}

func (c *Context) File(filePath string) {
	http.ServeFile(c.W, c.R, filePath)
}

// FileAttachment 自定义文件名的文件下载
func (c *Context) FileAttachment(filepath, filename string) {
	if isASCII(filename) {
		c.W.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	} else {
		c.W.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	c.File(filepath)
}

// FileFromFS 从文件系统获取
func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		c.R.URL.Path = old
	}(c.R.URL.Path)

	c.R.URL.Path = filepath

	http.FileServer(fs).ServeHTTP(c.W, c.R)
}

// Render 通用渲染
func (c *Context) Render(r render.Render, statusCode int) error {
	r.WriteContentType(c.W)
	c.W.WriteHeader(statusCode)
	return r.Render(c.W)
}
