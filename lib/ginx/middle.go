package ginx

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func BombRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				switch t := err.(type) {
				case HTTPError:
					c.String(t.Code, t.Message)
				case error:
					c.String(500, t.Error())
				case string:
					c.String(500, t)
				default:
					c.String(500, fmt.Sprintf("%v", t))
				}
			}
		}()
		c.Next()
	}
}

type HTTPError struct {
	Message string
	Code    int
}

func (p HTTPError) Error() string {
	return p.Message
}

func (p HTTPError) String() string {
	return p.Message
}

func Bomb(code int, format string, a ...interface{}) {
	panic(HTTPError{Code: code, Message: fmt.Sprintf(format, a...)})
}

func Dangerous(v interface{}, code ...int) {
	if v == nil {
		return
	}

	c := 200
	if len(code) > 0 {
		c = code[0]
	}

	switch t := v.(type) {
	case string:
		if t != "" {
			panic(HTTPError{Code: c, Message: t})
		}
	case error:
		panic(HTTPError{Code: c, Message: t.Error()})
	}
}
