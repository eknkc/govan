package govan

import (
	"bufio"
	"net"
	"net/http"
)

type (
	ResponseWriter interface {
		http.ResponseWriter
		http.Hijacker
		http.Flusher
		http.CloseNotifier

		Size() int
		Status() int
		Written() bool
		HeaderWritten() bool
		Hijacked() bool
		Close()
	}

	responseWriter struct {
		http.ResponseWriter
		status        int
		size          int
		headerWritten bool
		hijacked      bool
	}
)

func (w *responseWriter) reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.size = 0
}

func (w *responseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.status = code
	w.headerWritten = true
}

func (w *responseWriter) Write(data []byte) (n int, err error) {
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
	}

	n, err = w.ResponseWriter.Write(data)
	w.size += n
	return
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) HeaderWritten() bool {
	return w.headerWritten
}

func (w *responseWriter) Hijacked() bool {
	return w.hijacked
}

func (w *responseWriter) Written() bool {
	return w.size != 0
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijacked = true
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *responseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (w *responseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *responseWriter) Close() {

}
