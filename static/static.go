package static

import (
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ofunc/ws"
)

var allow = strings.Join([]string{http.MethodOptions, http.MethodGet, http.MethodHead}, ", ")

// New creates a static file middleware.
func New(root string, age int, exts ...string) func(*ws.Context) error {
	maxage := "max-age=" + strconv.Itoa(age)
	mgzip := map[string]bool{
		".html": true,
		".css":  true,
		".js":   true,
	}
	for _, ext := range exts {
		mgzip[ext] = true
	}

	return func(ctx *ws.Context) error {
		oerr := ctx.Next()
		if serr, ok := oerr.(*ws.StatusError); (ok && serr.Code() != http.StatusNotFound) || (!ok && !os.IsNotExist(oerr)) {
			return oerr
		}

		method := ctx.Request.Method
		if method == http.MethodOptions {
			ctx.ResponseWriter.Header().Add("Allow", allow)
			return ws.Status(http.StatusOK, "")
		}
		if method != http.MethodGet && method != http.MethodHead {
			return oerr
		}
		if strings.Contains(ctx.Path, "..") {
			return oerr
		}

		header := ctx.ResponseWriter.Header()
		if age > 0 {
			header.Set("Cache-Control", maxage)
		}
		path, isgzip, err := realpath(ctx.Request.Header, mgzip, root, ctx.Path)
		if err != nil {
			return oerr
		}
		if isgzip {
			header.Set("Content-Encoding", "gzip")
		}
		return ctx.File(path)
	}
}

func realpath(header http.Header, mgzip map[string]bool, root, path string) (string, bool, error) {
	path = filepath.FromSlash(path)
	fpath := filepath.Join(root, path)
	finfo, err := os.Stat(fpath)
	if err != nil {
		return "", false, err
	}
	if finfo.IsDir() {
		path = filepath.Join(path, "index.html")
		fpath = filepath.Join(fpath, "index.html")
		if finfo, err = os.Stat(fpath); err != nil {
			return "", false, err
		}
	}

	if finfo.Size() < 1024 || !mgzip[filepath.Ext(path)] || !strings.Contains(header.Get("Accept-Encoding"), "gzip") {
		return fpath, false, nil
	}

	gpath := filepath.Join(root, ".gzip", path)
	ginfo, err := os.Stat(gpath)
	if err != nil || ginfo.ModTime().Before(finfo.ModTime()) {
		err = compress(fpath, gpath)
	}
	return gpath, true, err
}

func compress(fpath, gpath string) error {
	if err := os.MkdirAll(filepath.Dir(gpath), os.ModePerm); err != nil {
		return err
	}

	r, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(gpath)
	if err != nil {
		return err
	}
	defer w.Close()

	zw, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		return err
	}
	defer zw.Close()

	if _, err = io.Copy(zw, r); err != nil {
		return err
	}
	if err := zw.Flush(); err != nil {
		return err
	}
	return w.Sync()
}
