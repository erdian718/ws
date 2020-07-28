package ws

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"strings"
	"time"
)

// Context represents the context which hold the HTTP request and response.
type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter

	app      *App
	status   int
	data     map[string]interface{}
	params   map[string]string
	handlers [][]func(*Context) error
	index    int
}

// Get gets the context data.
func (a *Context) Get(key string) interface{} {
	return a.data[key]
}

// Set sets the context data.
func (a *Context) Set(key string, value interface{}) {
	a.data[key] = value
}

// Param returns the parameter by key.
func (a *Context) Param(key string) (string, error) {
	if p, ok := a.params[key]; ok {
		return p, nil
	}
	if a.Request.Form == nil {
		a.Request.ParseMultipartForm(a.app.maxMemory)
	}
	if ps := a.Request.Form[key]; len(ps) > 0 {
		return ps[0], nil
	}
	return "", ErrMissingParam
}

// ParamJSON returns the JSON parameter by key.
func (a *Context) ParamJSON(key string, v interface{}) error {
	p, err := a.Param(key)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(p), v); err != nil {
		return a.BadRequest(err)
	}
	return nil
}

// ParamXML returns the XML parameter by key.
func (a *Context) ParamXML(key string, v interface{}) error {
	p, err := a.Param(key)
	if err != nil {
		return err
	}
	if err := xml.Unmarshal([]byte(p), v); err != nil {
		return a.BadRequest(err)
	}
	return nil
}

// ParamFile returns the file parameter by key.
func (a *Context) ParamFile(key string) (multipart.File, *multipart.FileHeader, error) {
	if a.Request.MultipartForm == nil {
		if err := a.Request.ParseMultipartForm(a.app.maxMemory); err != nil {
			return nil, nil, a.BadRequest(err)
		}
	}
	if a.Request.MultipartForm != nil && a.Request.MultipartForm.File != nil {
		if fhs := a.Request.MultipartForm.File[key]; len(fhs) > 0 {
			f, err := fhs[0].Open()
			if err != nil {
				return nil, nil, a.BadRequest(err)
			}
			return f, fhs[0], nil
		}
	}
	return nil, nil, ErrMissingParam
}

// SetStatus sets the status code.
func (a *Context) SetStatus(code int) *Context {
	a.status = code
	return a
}

// Status responses the status code.
func (a *Context) Status(code int) error {
	a.ResponseWriter.WriteHeader(code)
	return nil
}

// BadRequest wraps ErrBadRequest with err.
func (a *Context) BadRequest(err error) error {
	return fmt.Errorf("%w: %v", ErrBadRequest, err)
}

// NotFound wraps ErrNotFound with err.
func (a *Context) NotFound(err error) error {
	return fmt.Errorf("%w: %v", ErrNotFound, err)
}

// Text responses the text content.
func (a *Context) Text(v string) error {
	a.ResponseWriter.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	a.writeHeader()
	a.ResponseWriter.Write([]byte(v))
	return nil
}

// JSON responses the JSON content.
func (a *Context) JSON(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	a.ResponseWriter.Header().Set("Content-Type", "application/json; charset=UTF-8")
	a.writeHeader()
	a.ResponseWriter.Write(b)
	return nil
}

// XML responses the XML content.
func (a *Context) XML(v interface{}) error {
	b, err := xml.Marshal(v)
	if err != nil {
		return err
	}
	a.ResponseWriter.Header().Set("Content-Type", "application/xml; charset=UTF-8")
	a.writeHeader()
	a.ResponseWriter.Write(b)
	return nil
}

// File responses the file content.
func (a *Context) File(name string) error {
	return sendFile(a.ResponseWriter, a.Request, name)
}

// Content responses the content.
func (a *Context) Content(name string, modtime time.Time, content io.ReadSeeker) error {
	http.ServeContent(a.ResponseWriter, a.Request, name, modtime, content)
	return nil
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

// Next calls the next handler.
func (a *Context) Next() error {
	if len(a.handlers) <= 0 {
		return nil
	}
	hs := a.handlers[0]
	if a.index >= len(hs) {
		a.index = 0
		a.handlers = a.handlers[1:]
		return a.Next()
	}

	return hs[a.index](&Context{
		Request:        a.Request,
		ResponseWriter: a.ResponseWriter,
		app:            a.app,
		data:           a.data,
		params:         a.params,
		handlers:       a.handlers,
		index:          a.index + 1,
	})
}

func (a *Context) writeHeader() {
	if a.status == 0 {
		switch a.Request.Method {
		case "POST":
			a.status = http.StatusCreated
		case "DELETE":
			a.status = http.StatusNoContent
		default:
			a.status = http.StatusOK
		}
	}
	a.ResponseWriter.WriteHeader(a.status)
}
