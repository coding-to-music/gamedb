package middleware

import (
	"net/http"
)

type ResponseWriterWithLength struct {
	http.ResponseWriter
	length int
}

func (w *ResponseWriterWithLength) Write(b []byte) (n int, err error) {

	n, err = w.ResponseWriter.Write(b)
	w.length += n

	return n, err
}

func (w *ResponseWriterWithLength) Length() int {
	return w.length
}

func ResponseWriterWithLengthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rec := ResponseWriterWithLength{w, 0}
		next.ServeHTTP(&rec, r)
	})
}
