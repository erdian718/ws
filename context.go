package ws

import (
	"/net/http"
	"encoding/json"
	"encoding/xml"
	"io"
	"mime/multipart"
	"time"
)

// Context represents the context which hold the HTTP request and response.
type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter

	group  *Group
	params map[string]string
}

// Param returns the parameter by key.
func (a *Context) Param(key string) (string, error) {
	if p, ok := a.params[key]; ok {
		return p, nil
	}
	if a.Request.Form == nil {
		if a.Request.ParseMultipartForm(a.group.maxFormSize) != nil {
			return "", ErrBadRequest
		}
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
	if json.Unmarshal([]byte(p), v) != nil {
		return ErrBadRequest
	}
	return nil
}

// ParamXML returns the XML parameter by key.
func (a *Context) ParamXML(key string, v interface{}) error {
	p, err := a.Param(key)
	if err != nil {
		return err
	}
	if xml.Unmarshal([]byte(p), v) != nil {
		return ErrBadRequest
	}
	return nil
}

// ParamFile returns the file parameter by key.
func (a *Context) ParamFile(key string) (multipart.File, *multipart.FileHeader, error) {
	if a.Request.MultipartForm == nil {
		if a.Request.ParseMultipartForm(a.group.maxFormSize) != nil {
			return nil, nil, ErrBadRequest
		}
	}
	if a.Request.MultipartForm != nil && a.Request.MultipartForm.File != nil {
		if fhs := a.Request.MultipartForm.File[key]; len(fhs) > 0 {
			f, err := fhs[0].Open()
			if err != nil {
				return nil, nil, ErrBadRequest
			}
			return f, fhs[0], nil
		}
	}
	return nil, nil, ErrMissingParam
}

// Status responses the status code.
func (a *Context) Status(code int) error {
	a.ResponseWriter.WriteHeader(code)
	return nil
}

// Text responses the text content.
func (a *Context) Text(v string) error {
	a.ResponseWriter.Header().Set("Content-Type", "text/plain; charset=UTF-8")
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
	a.ResponseWriter.Write(b)
	return nil
}

// File responses the file content.
func (a *Context) File(name string) error {
	http.ServeFile(a.ResponseWriter, a.Request, name)
	return nil
}

// Content responses the content.
func (a *Context) Content(name string, modtime time.Time, content io.ReadSeeker) error {
	http.ServeContent(a.ResponseWriter, a.Request, name, modtime, content)
	return nil
}
