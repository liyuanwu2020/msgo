package msgo

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
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

func (c *Context) JSON(status int, data any) error {
	c.W.Header().Set("Content-Type", "application/json;charset=utf-8")
	c.W.WriteHeader(status)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = c.W.Write(jsonData)
	return err
}

func (c *Context) XML(status int, data any) error {
	c.W.Header().Set("Content-Type", "application/xml;charset=utf-8")
	c.W.WriteHeader(status)
	//xmlData, err := xml.Marshal(data)
	//if err != nil {
	//	return err
	//}
	//_, err = c.W.Write(xmlData)
	err := xml.NewEncoder(c.W).Encode(data)
	return err
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

// Redirect 重定向
func (c *Context) Redirect(status int, location string) {
	if (status < http.StatusMultipleChoices || status > http.StatusPermanentRedirect) && status != http.StatusCreated {
		panic(fmt.Sprintf("cannot redirect with status code %d", status))
	}
	http.Redirect(c.W, c.R, location, status)
}

// String 字符串
func (c *Context) String(status int, format string, values ...any) error {
	c.W.Header().Set("Content-Type", "text/plain;charset=utf-8")
	c.W.WriteHeader(status)
	var err error
	if len(values) > 0 {
		_, err = fmt.Fprintf(c.W, format, values...)
	} else {
		_, err = c.W.Write(StringToBytes(format))
	}
	return err
}
