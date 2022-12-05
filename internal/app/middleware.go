package app

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"
)

//Сервис поддерживает gzip
type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

//middleware обработчик подменяет writer на gzip.writer
//если клиент принимает сжатые ответы
func gzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// проверяем, что клиент поддерживает gzip-сжатие
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// если gzip не поддерживается, передаём управление
			// дальше без изменений
			next.ServeHTTP(w, r)
			return
		}

		// создаём gzip.Writer поверх текущего w
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		//передаём обработчику страницы переменную типа gzipWriter для вывода данных
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

type gzipReaderCloser struct {
	io.ReadCloser
	Reader io.Reader
}

func (r gzipReaderCloser) Read(b []byte) (int, error) {
	return r.Reader.Read(b)
}

//middleware обработчик подменяет writer на gzip.writer
//если клиент принимает сжатые ответы
func ungzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// проверяем, что клиент поддерживает gzip-сжатие
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			// если gzip не поддерживается, передаём управление
			// дальше без изменений
			next.ServeHTTP(w, r)
			return
		}

		// создаём gzip.Reader поверх текущего Body
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		r.Body = gzipReaderCloser{Reader: gz, ReadCloser: r.Body}
		//передаём обработчику страницы переменную типа gzipWriter для вывода данных
		next.ServeHTTP(w, r)
	})
}

func (c *cookieService) Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := c.checkSign(r, "userID")
		if err == nil {
			next.ServeHTTP(w, r)
			return
		}

		if errors.Is(err, http.ErrNoCookie) || errors.Is(err, ErrInvalidValue) {
			err = c.createAndSign(w)
			if err == nil {
				next.ServeHTTP(w, r)
				return
			}
		}

		io.WriteString(w, err.Error())
		return
	})
}
