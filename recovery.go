package govan

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
)

func NewRecovery() Handler {
	return func(l *log.Logger, rw http.ResponseWriter, r *http.Request, c *Ctx) {
		defer func() {
			if err := recover(); err != nil {
				if rw.Header().Get("Content-Type") == "" {
					rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
				}

				rw.WriteHeader(http.StatusInternalServerError)

				stack := make([]byte, 1024*8)
				stack = stack[:runtime.Stack(stack, false)]

				f := "PANIC: %s\n%s"
				l.Printf(f, err, stack)

				fmt.Fprintf(rw, f, err, stack)
			}
		}()

		c.Next()
	}
}
