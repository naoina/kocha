package kocha_test

import (
	"net/http"
	"testing"

	"github.com/naoina/kocha"
)

type nullResponseWriter struct {
	header http.Header
}

func newNullResponseWriter() *nullResponseWriter {
	return &nullResponseWriter{
		header: make(http.Header),
	}
}

func (w *nullResponseWriter) Header() http.Header {
	return w.header
}

func (w *nullResponseWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (w *nullResponseWriter) WriteHeader(n int) {
}

func BenchmarkServeHTTP(b *testing.B) {
	app := kocha.NewTestApp()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			b.Fatal(err)
		}
		w := newNullResponseWriter()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkNew(b *testing.B) {
	app := kocha.NewTestApp()
	config := app.Config
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := kocha.New(config); err != nil {
			b.Fatal(err)
		}
	}
}
