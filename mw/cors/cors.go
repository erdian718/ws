package cors

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/ofunc/ws"
)

// Control is the access control.
type Control struct {
	Origin           string
	ExposeHeaders    []string
	AllowHeaders     []string
	AllowMethods     []string
	AllowCredentials bool
	MaxAge           int
}

// New creates a cors middleware.
func New(cs ...*Control) func(*ws.Context) error {
	csm := make(map[string]*Control)
	for _, c := range cs {
		csm[c.Origin] = c
	}
	return func(ctx *ws.Context) error {
		method := ctx.Request.Method
		path := ctx.Request.URL.Path
		reqheader := ctx.Request.Header
		resheader := ctx.ResponseWriter.Header()

		origin := reqheader.Get("Origin")
		if origin == "" {
			return ctx.Next()
		}

		c, ok := csm[origin]
		if !ok {
			return ws.Status(http.StatusForbidden, method+" "+path+" from "+origin)
		}
		ok = false
		if method != http.MethodOptions {
			for _, m := range c.AllowMethods {
				if m == method {
					ok = true
					break
				}
			}
		}
		if !ok {
			return ws.Status(http.StatusForbidden, method+" "+path+" from "+origin)
		}

		resheader.Set("Access-Control-Allow-Origin", c.Origin)
		if len(c.ExposeHeaders) > 0 {
			resheader.Set("Access-Control-Expose-Headers", strings.Join(c.ExposeHeaders, ", "))
		}

		err := ctx.Next()
		if method == http.MethodOptions {
			if c.AllowCredentials {
				resheader.Set("Access-Control-Allow-Credentials", "true")
			}
			if len(c.AllowHeaders) > 0 {
				resheader.Set("Access-Control-Allow-Headers", strings.Join(c.AllowHeaders, ", "))
			}
			if len(c.AllowMethods) > 0 {
				resheader.Set("Access-Control-Allow-Methods", strings.Join(c.AllowMethods, ", "))
			} else {
				resheader.Set("Access-Control-Allow-Methods", resheader.Get("Allow"))
			}
			if c.MaxAge > 0 {
				resheader.Set("Access-Control-Max-Age", strconv.Itoa(c.MaxAge))
			}
		}
		return err
	}
}
