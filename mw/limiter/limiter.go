package limiter

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ofunc/ws"
)

type info struct {
	dur  float64
	time time.Time
}

// New creates a limiter middleware.
func New(size int64, freq int, timeout time.Duration) func(*ws.Context) error {
	var mdur float64
	var minfo map[string]info
	var mutex sync.Mutex
	if freq > 0 {
		mdur = 1 / float64(freq)
		minfo = make(map[string]info)
	}

	return func(ctx *ws.Context) error {
		if size > 0 {
			ctx.Request.Body = http.MaxBytesReader(ctx.ResponseWriter, ctx.Request.Body, size)
		}
		if freq > 0 {
			now, key, dur := time.Now(), ctx.RealIP(), 2*mdur
			mutex.Lock()
			if info, ok := minfo[key]; ok {
				dur = 0.01*now.Sub(info.time).Seconds() + 0.99*info.dur
			}
			minfo[key] = info{
				dur:  dur,
				time: now,
			}
			mutex.Unlock()
			if dur < mdur {
				err := ws.Status(http.StatusTooManyRequests, key)
				log.Println(err)
				return err
			}
		}
		if timeout <= 0 {
			return ctx.Next()
		}

		ch := make(chan error)
		go func() {
			defer func() {
				if x := recover(); x != nil {
					ch <- fmt.Errorf("ws: %v", x)
				}
				close(ch)
			}()
			ch <- ctx.Next()
		}()

		select {
		case err := <-ch:
			return err
		case <-time.After(timeout):
			return ws.Status(http.StatusRequestTimeout, ctx.RealIP())
		}
	}
}
