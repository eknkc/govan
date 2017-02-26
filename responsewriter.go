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
		Written() bool
		HeaderWritten() bool
	}

	responseWriter struct {
		http.ResponseWriter
		size          int
		headerWritten bool
	}
)

func (w *responseWriter) reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.size = 0
}

func (w *responseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.headerWritten = true
}

func (w *responseWriter) Write(data []byte) (n int, err error) {
	n, err = w.ResponseWriter.Write(data)
	w.size += n
	return
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) HeaderWritten() bool {
	return w.headerWritten
}

func (w *responseWriter) Written() bool {
	return w.size != 0
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *responseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (w *responseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}
