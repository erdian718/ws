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
	"path/filepath"
	"strings"
	"time"
)

// Context represents the context which hold the HTTP request and response.
type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Path           string

	code   int
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
		method := a.Request.Method
		if method == http.MethodOptions {
			allow := []string{http.MethodOptions}
			for m := range a.router.handlers {
				allow = append(allow, m)
				if m == http.MethodGet {
					allow = append(allow, http.MethodHead)
				}
			}
			a.ResponseWriter.Header().Add("Allow", strings.Join(allow, ", "))
			a.ResponseWriter.WriteHeader(http.StatusOK)
			a.ResponseWriter.Write([]byte(""))
			return nil
		}
		if method == http.MethodHead {
			method = http.MethodGet
		}

		hs, ok := a.router.handlers[method]
		if !ok {
			return Status(http.StatusMethodNotAllowed, method+" "+a.Request.URL.Path)
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

// Status sets the status code.
func (a *Context) Status(code int) {
	a.code = code
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
	f, fh, err := a.Request.FormFile(key)
	if err != nil {
		err = Status(http.StatusBadRequest, err.Error())
	}
	return f, fh, err
}

// ParseJSON parse the JSON data.
func (a *Context) ParseJSON(value interface{}) error {
	err := json.NewDecoder(a.Request.Body).Decode(value)
	if err != nil {
		err = Status(http.StatusBadRequest, err.Error())
	}
	return err
}

// ParseXML parse the XML data.
func (a *Context) ParseXML(value interface{}) error {
	err := xml.NewDecoder(a.Request.Body).Decode(value)
	if err != nil {
		err = Status(http.StatusBadRequest, err.Error())
	}
	return err
}

// Text responses the text content.
func (a *Context) Text(value string) error {
	a.ResponseWriter.Header().Set("Content-Type", "text/plain; charset=utf-8")
	a.ResponseWriter.WriteHeader(a.statusCode())
	a.ResponseWriter.Write([]byte(value))
	return nil
}

// JSON responses the JSON content.
func (a *Context) JSON(value interface{}) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	a.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	a.ResponseWriter.WriteHeader(a.statusCode())
	a.ResponseWriter.Write(b)
	return nil
}

// XML responses the XML content.
func (a *Context) XML(value interface{}) error {
	b, err := xml.Marshal(value)
	if err != nil {
		return err
	}
	a.ResponseWriter.Header().Set("Content-Type", "application/xml; charset=utf-8")
	a.ResponseWriter.WriteHeader(a.statusCode())
	a.ResponseWriter.Write(b)
	return nil
}

// Content responses the content.
func (a *Context) Content(name string, modtime time.Time, content io.ReadSeeker) error {
	http.ServeContent(a.ResponseWriter, a.Request, name, modtime, content)
	return nil
}

// File responses the file content.
func (a *Context) File(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Status(http.StatusNotFound, err.Error())
		}
		if os.IsPermission(err) {
			return Status(http.StatusForbidden, err.Error())
		}
		return err
	}
	if stat.IsDir() {
		path = filepath.Join(path, "index.html")
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err = f.Stat()
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

func (a *Context) statusCode() int {
	if a.code > 0 {
		return a.code
	}
	switch a.Request.Method {
	case http.MethodPost:
		return http.StatusCreated
	case http.MethodDelete:
		return http.StatusNoContent
	default:
		return http.StatusOK
	}
}
