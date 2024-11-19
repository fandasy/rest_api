package req_controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"restApi/internal/config"
	"sync"
	"sync/atomic"
	"time"
)

type ReqCounter struct {
	m  sync.Map
	rl LimitOptions
}

type UserControl struct {
	msgCounter  int32
	lastMsgTime int64 // хранить время в наносекундах
	bannedUntil int64 // хранить время в наносекундах
}

type LimitOptions struct {
	limit    uint
	interval time.Duration
	banTime  time.Duration
}

func New(limit config.ReqLimit) *ReqCounter {
	return &ReqCounter{
		rl: LimitOptions{
			limit:    limit.MaxNumReq,
			interval: limit.TimeSlice,
			banTime:  limit.BanTime,
		},
	}
}

func (rc *ReqCounter) Checking() gin.HandlerFunc {
	fn := func(c *gin.Context) {
		clientIP := c.ClientIP()

		if ok := rc.processor(clientIP); !ok {
			http.Error(c.Writer, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		c.Next()
	}

	return fn
}

func (rc *ReqCounter) processor(username string) bool {
	userInfo, loaded := rc.m.LoadOrStore(username, &UserControl{
		msgCounter:  1,
		lastMsgTime: time.Now().UnixNano(),
	})
	if !loaded {
		return true
	}

	user := userInfo.(*UserControl)

	bannedUntil := atomic.LoadInt64(&user.bannedUntil)
	if bannedUntil > time.Now().UnixNano() {
		return false
	}

	lastMsgTime := atomic.LoadInt64(&user.lastMsgTime)
	if time.Since(time.Unix(0, lastMsgTime)) < rc.rl.interval {

		if atomic.LoadInt32(&user.msgCounter) >= int32(rc.rl.limit) {

			user.Ban(time.Now().Add(rc.rl.banTime))

			return false
		}

		user.Add(1)

	} else {
		user.Reset()
	}

	return true
}

func (u *UserControl) Add(number uint) {
	atomic.AddInt32(&u.msgCounter, int32(number))
}

func (u *UserControl) Reset() {
	atomic.StoreInt32(&u.msgCounter, 1)
	atomic.StoreInt64(&u.lastMsgTime, time.Now().UnixNano())
}

func (u *UserControl) Ban(bannedUntil time.Time) {
	atomic.StoreInt32(&u.msgCounter, 0)
	atomic.StoreInt64(&u.bannedUntil, bannedUntil.UnixNano())
}
