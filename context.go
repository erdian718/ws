package ws

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Context represents the context which hold the HTTP request and response.
type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Path           string

	datas  map[string]interface{}
	params map[string]string
	querys url.Values
	router *Router
	index  int
}

// Next calls the next handler.
func (a *Context) Next() error {
	if a.index < 0 {
		ms := a.router.middlewares
		return ms[a.index+len(ms)](&Context{
			Request:        a.Request,
			ResponseWriter: a.ResponseWriter,
			Path:           a.Path,

			datas:  a.datas,
			params: a.params,
			querys: a.querys,
			router: a.router,
			index:  a.index + 1,
		})
	}

	if a.Path == "" {
		hs, ok := a.router.handlers[a.Request.Method]
		if !ok {
			return Status(http.StatusMethodNotAllowed, a.Request.Method+" "+a.Request.URL.Path)
		}
		if a.index >= len(hs) {
			return nil
		}
		return hs[a.index](&Context{
			Request:        a.Request,
			ResponseWriter: a.ResponseWriter,
			Path:           a.Path,

			datas:  a.datas,
			params: a.params,
			querys: a.querys,
			router: a.router,
			index:  a.index + 1,
		})
	}

	router, path, param := a.router.Match(a.Path)
	if router == nil {
		return Status(http.StatusNotFound, a.Request.URL.Path)
	}
	if key := a.router.key; key != "" {
		a.params[key[1:]] = param
	}

	return (&Context{
		Request:        a.Request,
		ResponseWriter: a.ResponseWriter,
		Path:           path,

		datas:  a.datas,
		params: a.params,
		querys: a.querys,
		router: router,
		index:  -len(router.middlewares),
	}).Next()
}

// Get gets the context data.
func (a *Context) Get(key string) interface{} {
	return a.datas[key]
}

// Set sets the context data.
func (a *Context) Set(key string, value interface{}) {
	a.datas[key] = value
}

// Param return the param by key.
func (a *Context) Param(key string) string {
	return a.params[key]
}

// Query returns the first value associated with the given key.
func (a *Context) Query(key string) string {
	if a.querys == nil {
		a.querys, _ = url.ParseQuery(a.Request.URL.RawQuery)
	}
	return a.querys.Get(key)
}

// FormValue returns the first value for the named component of the request body.
func (a *Context) FormValue(key string) string {
	return a.Request.PostFormValue(key)
}

// FormFile returns the first file for the provided form key.
func (a *Context) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return a.Request.FormFile(key)
}

// ParseJSON parse the JSON data.
func (a *Context) ParseJSON(value interface{}) error {
	err := json.NewDecoder(a.Request.Body).Decode(value)
	if err != nil {
		err = Status(http.StatusBadRequest, err)
	}
	return err
}

// ParseXML parse the XML data.
func (a *Context) ParseXML(value interface{}) error {
	err := xml.NewDecoder(a.Request.Body).Decode(value)
	if err != nil {
		err = Status(http.StatusBadRequest, err)
	}
	return err
}

// Text responses the text content.
func (a *Context) Text(code int, value string) error {
	a.ResponseWriter.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	a.ResponseWriter.WriteHeader(a.statusCode(code))
	a.ResponseWriter.Write([]byte(value))
	return nil
}

// JSON responses the JSON content.
func (a *Context) JSON(code int, value interface{}) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	a.ResponseWriter.Header().Set("Content-Type", "application/json; charset=UTF-8")
	a.ResponseWriter.WriteHeader(a.statusCode(code))
	a.ResponseWriter.Write(b)
	return nil
}

// XML responses the XML content.
func (a *Context) XML(code int, value interface{}) error {
	b, err := xml.Marshal(value)
	if err != nil {
		return err
	}
	a.ResponseWriter.Header().Set("Content-Type", "application/xml; charset=UTF-8")
	a.ResponseWriter.WriteHeader(a.statusCode(code))
	a.ResponseWriter.Write(b)
	return nil
}

// Content responses the content.
func (a *Context) Content(name string, modtime time.Time, content io.ReadSeeker) error {
	http.ServeContent(a.ResponseWriter, a.Request, name, modtime, content)
	return nil
}

// File responses the file content.
func (a *Context) File(name string) error {
	f, err := os.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			err = Status(http.StatusNotFound, err)
		}
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}
	return a.Content(f.Name(), stat.ModTime(), f)
}

// RealIP returns the real client IP.
func (a *Context) RealIP() string {
	header := a.Request.Header
	if ip := header.Get("X-Forwarded-For"); ip != "" {
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	}
	if ip := header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	ra, _, _ := net.SplitHostPort(a.Request.RemoteAddr)
	return ra
}

func (a *Context) statusCode(code int) int {
	if code > 0 {
		return code
	}
	switch a.Request.Method {
	case "POST":
		return http.StatusCreated
	case "DELETE":
		return http.StatusNoContent
	default:
		return http.StatusOK
	}
}
