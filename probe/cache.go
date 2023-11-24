package probe

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var (
	c = cache.New(5*time.Second, time.Minute)
)

func CacheGetBytes(key string) []byte {
	if x, found := c.Get(key); found {
		return x.([]byte)
	}
	return nil
}

func CacheSetBytes(key string, bs []byte, d ...time.Duration) {
	if len(d) > 0 {
		c.Set(key, bs, d[0])
		return
	}

	c.Set(key, bs, cache.DefaultExpiration)
}
