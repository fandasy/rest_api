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
	rl RateLimit
}

type UserControl struct {
	msgCounter  uint32
	lastMsgTime uint64 // хранить время в наносекундах
	bannedUntil uint64 // хранить время в наносекундах
}

type RateLimit struct {
	limit    uint32
	interval time.Duration
	banTime  time.Duration
}

func New(limit config.ReqLimit) *ReqCounter {
	return &ReqCounter{
		rl: RateLimit{
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
		lastMsgTime: uint64(time.Now().UnixNano()),
	})
	if !loaded {
		return true
	}

	user := userInfo.(*UserControl)

	bannedUntil := atomic.LoadUint64(&user.bannedUntil)
	if bannedUntil > uint64(time.Now().UnixNano()) {
		return false
	}

	lastMsgTime := atomic.LoadUint64(&user.lastMsgTime)
	if time.Since(time.Unix(0, int64(lastMsgTime))) < rc.rl.interval {

		if atomic.LoadUint32(&user.msgCounter) >= rc.rl.limit {

			user.Ban(time.Now().Add(rc.rl.banTime))

			return false
		}

		user.Add(1)

	} else {
		user.Reset()
	}

	return true
}

func (u *UserControl) Add(number uint32) {
	atomic.AddUint32(&u.msgCounter, number)
}

func (u *UserControl) Reset() {
	atomic.StoreUint32(&u.msgCounter, 1)
	atomic.StoreUint64(&u.lastMsgTime, uint64(time.Now().UnixNano()))
}

func (u *UserControl) Ban(bannedUntil time.Time) {
	atomic.StoreUint32(&u.msgCounter, 0)
	atomic.StoreUint64(&u.bannedUntil, uint64(bannedUntil.UnixNano()))
}
