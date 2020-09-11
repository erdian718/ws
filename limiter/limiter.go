package limiter

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ofunc/ws"
)

const (
	k = 0.010
	m = 1.200
)

type info struct {
	dur  float64
	time time.Time
}

// New creates a limiter middleware.
func New(size int64, duration time.Duration, fkey func(*ws.Context) string) func(*ws.Context) error {
	var mdur float64
	var stime time.Time
	var sdur time.Duration
	var hinfo map[string]info
	var cinfo map[string]info
	var mutex sync.Mutex
	var pool *sync.Pool

	putInfoMap := func(m map[string]info) {
		if m == nil {
			return
		}
		for k := range m {
			delete(m, k)
		}
		pool.Put(m)
	}

	if duration > 0 {
		mdur = duration.Seconds()
		sdur = time.Duration(1e9 * mdur * (m + k - 1) / k)
		if sdur < 10*time.Minute {
			sdur = 10 * time.Minute
		}
		stime = time.Now().Add(sdur)

		pool = &sync.Pool{
			New: func() interface{} {
				return make(map[string]info)
			},
		}
		hinfo = pool.Get().(map[string]info)

		if fkey == nil {
			fkey = func(ctx *ws.Context) string {
				return ctx.RealIP()
			}
		}
	}

	return func(ctx *ws.Context) error {
		if size > 0 {
			ctx.Request.Body = http.MaxBytesReader(ctx.ResponseWriter, ctx.Request.Body, size)
		}
		if duration > 0 {
			now, key, dur := time.Now(), fkey(ctx), m*mdur
			mutex.Lock()
			x, ok := hinfo[key]
			if !ok {
				x, ok = cinfo[key]
			}
			if ok {
				dur = k*now.Sub(x.time).Seconds() + (1-k)*x.dur
			}

			if now.After(stime) {
				go putInfoMap(cinfo)
				stime, hinfo, cinfo = now.Add(sdur), pool.Get().(map[string]info), hinfo
			}
			hinfo[key] = info{
				dur:  dur,
				time: now,
			}
			mutex.Unlock()

			if dur < mdur {
				err := ws.Status(http.StatusTooManyRequests, key)
				if x.dur >= mdur {
					log.Println(err)
				}
				return err
			}
		}
		return ctx.Next()
	}
}
