package msgo

import (
	"encoding/json"
	"encoding/xml"
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
