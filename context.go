package msgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/liyuanwu2020/msgo/render"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
)

const defaultMultipartMemory = 32 << 20

type Context struct {
	W                     http.ResponseWriter
	R                     *http.Request
	NodeRouterName        string
	RequestMethod         string
	engine                *Engine
	queryCache            url.Values
	DisallowUnknownFields bool
	IsValidate            bool
	requiredTag           string
}

func (c *Context) DealJson(data any) error {
	body := c.R.Body
	if c.R == nil || body == nil {
		return errors.New("invalid json request")
	}
	decoder := json.NewDecoder(body)
	if c.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if c.IsValidate {
		err := c.validateParam(data, decoder)
		if err != nil {
			return err
		}
	} else {
		err := decoder.Decode(data)
		if err != nil {
			return err
		}
	}
	log.Println("github validator begin")
	return validate(data)
}

func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dstName string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer func(src multipart.File) {
		err := src.Close()
		if err != nil {

		}
	}(src)
	dst, err := os.Create(dstName)
	if err != nil {
		return err
	}
	defer func(dst *os.File) {
		err := dst.Close()
		if err != nil {

		}
	}(dst)

	_, err = io.Copy(dst, src)
	return err
}

// FormFile 获取文件
func (c *Context) FormFile(key string) (*multipart.FileHeader, error) {
	file, header, err := c.R.FormFile(key)
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)
	return header, err
}

func (c *Context) FormFiles(key string) []*multipart.FileHeader {
	multipartForm, err := c.MultipartForm()
	if err != nil {
		return make([]*multipart.FileHeader, 0)
	}
	return multipartForm.File[key]
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.R.ParseMultipartForm(defaultMultipartMemory)
	return c.R.MultipartForm, err
}

func (c *Context) GetPost(key string) (string, error) {
	if err := c.R.ParseMultipartForm(defaultMultipartMemory); err != nil {
		if !errors.Is(err, http.ErrNotMultipart) {
			return "", err
		}
	}
	return c.R.PostForm.Get(key), nil
}

func (c *Context) GetAllPost() (url.Values, error) {
	if err := c.R.ParseMultipartForm(defaultMultipartMemory); err != nil {
		if !errors.Is(err, http.ErrNotMultipart) {
			return nil, err
		}
	}
	return c.R.PostForm, nil
}

// GetMapQuery http://localhost:8080/queryMap?user[id]=1&user[name]=张三
func (c *Context) GetMapQuery(key string) (map[string]string, bool) {
	c.initQueryCache()
	rs := make(map[string]string)
	exists := false
	for k, v := range c.queryCache {
		start := strings.Index(k, "[")

		if start > 0 && k[:start] == key {
			end := strings.Index(k[start+1:], "]")
			if end > 0 {
				exists = true
				rs[k[start+1:][:end]] = v[0]
			}

		}
	}
	return rs, exists
}

func (c *Context) DefaultQuery(key, defaultValue string) string {
	array, ok := c.GetQueryArray(key)
	if !ok {
		return defaultValue
	}
	return array[0]
}

func (c *Context) GetQuery(key string) string {
	c.initQueryCache()
	return c.queryCache.Get(key)
}

func (c *Context) QueryArray(key string) (values []string) {
	c.initQueryCache()
	values, _ = c.queryCache[key]
	return
}

func (c *Context) GetQueryArray(key string) (values []string, ok bool) {
	c.initQueryCache()
	values, ok = c.queryCache[key]
	return
}

func (c *Context) initQueryCache() {
	//if c.queryCache == nil {
	if c.R != nil {
		c.queryCache = c.R.URL.Query()
	} else {
		c.queryCache = url.Values{}
	}
	//}
	log.Println(c.queryCache)
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
	if statusCode != http.StatusOK {
		c.W.WriteHeader(statusCode)
	}
	return r.Render(c.W)
}

func (c *Context) validateParam(data any, decoder *json.Decoder) error {
	valueOf := reflect.ValueOf(data)
	if valueOf.Kind() != reflect.Pointer {
		return errors.New("data argument must have a pointer type")
	}
	elem := valueOf.Elem().Interface()
	value := reflect.ValueOf(elem)
	switch value.Kind() {
	case reflect.Struct:
		return checkParam(value, data, decoder)
	case reflect.Array, reflect.Slice:
		ele := value.Type().Elem()
		if ele.Kind() == reflect.Ptr {
			ele = ele.Elem()
		}
		if ele.Kind() == reflect.Struct {
			return checkSliceParam(ele, data, decoder)
		}
	default:
		return decoder.Decode(data)
	}
	return nil
}

func checkSliceParam(elem reflect.Type, data any, decoder *json.Decoder) error {
	mapVal := make([]map[string]interface{}, 0)
	//解析为map
	err := decoder.Decode(&mapVal)
	if err != nil {
		return err
	}
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		name := field.Name
		jsonName := field.Tag.Get("json")
		requiredTag := field.Tag.Get("required")
		if jsonName == "" {
			jsonName = name
		}
		for i, v := range mapVal {
			if v[jsonName] == nil && requiredTag != "" {
				return errors.New(fmt.Sprintf("row[%d] field [%s] is not exist", i+1, jsonName))
			}
		}

	}
	b, _ := json.Marshal(mapVal)
	return json.Unmarshal(b, data)
}

func checkParam(value reflect.Value, data any, decoder *json.Decoder) error {
	mapVal := make(map[string]interface{})
	//解析为map
	err := decoder.Decode(&mapVal)
	if err != nil {
		return err
	}
	for i := 0; i < value.NumField(); i++ {
		field := value.Type().Field(i)
		name := field.Name
		jsonName := field.Tag.Get("json")
		requiredTag := field.Tag.Get("required")
		if jsonName == "" {
			jsonName = name
		}
		if mapVal[jsonName] == nil && requiredTag != "" {
			return errors.New(fmt.Sprintf("field [%s] is not exist", jsonName))
		}
	}
	b, _ := json.Marshal(mapVal)
	return json.Unmarshal(b, data)
}

type SliceValidationError []error

func (err SliceValidationError) Error() string {
	n := len(err)
	switch n {
	case 0:
		return ""
	default:
		var b strings.Builder
		if err[0] != nil {
			_, err := fmt.Fprintf(&b, "[%d]: %s", 0, err[0].Error())
			if err != nil {
				return ""
			}
		}
		if n > 1 {
			for i := 1; i < n; i++ {
				if err[i] != nil {
					b.WriteString("\n")
					_, err := fmt.Fprintf(&b, "[%d]: %s", i, err[i].Error())
					if err != nil {
						return ""
					}
				}
			}
		}
		return b.String()
	}
}

func validate(data any) error {
	if data == nil {
		return nil
	}
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		count := v.Len()
		errs := make(SliceValidationError, 0)
		for i := 0; i < count; i++ {
			if err := validateStruct(v.Index(i).Interface()); err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) == 0 {
			return nil
		}
		return errs
	case reflect.Ptr:
		return validate(v.Elem().Interface())
	case reflect.Struct:
		return validateStruct(data)
	}
	return nil
}

func validateStruct(data any) error {
	return validator.New().Struct(data)
}
