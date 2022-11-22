package app

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	url2 "net/url"
	"strings"
)

//Эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор сокращённого URL и
//возвращает ответ с кодом 307 и оригинальным URL в HTTP-заголовке Location.
func (h *Handler) getHandler(w http.ResponseWriter, r *http.Request) {
	//ищем в хранилище соответсвующий полный юрл
	expandURL, ok := h.storage.LookUp(r.URL.Path)

	if !ok {
		http.Error(w, "Passed short url not found", http.StatusBadRequest)
		return
	}

	//формируем ответ
	w.Header().Add("Location", expandURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

//ндпоинт POST / принимает в теле запроса строку URL для сокращения
//и возвращает ответ с кодом 201 и сокращённым URL в виде текстовой строки в теле.
func (h *Handler) postHandler(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	url, err := url2.Parse(string(b))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//сокращаем юрл и добавляем его в хранилище
	//мапим только путь потому что префиксы у всех урлов одинаковые
	short, _ := h.shortenURL(url)

	h.storage.Insert(short.Path, string(b))
	//устанавливаем статус ответа
	w.WriteHeader(http.StatusCreated)
	//пишем в тело ответа сокращенный url
	_, _ = w.Write([]byte(short.String()))
}

type InMessage struct {
	ExpandURL url2.URL `json:"url"`
}

//рабоче-крестьянским методом валидируем урл при чтении
func (im *InMessage) UnmarshalJSON(data []byte) error {
	aliasValue := &struct {
		RawURL string `json:"url"`
	}{}
	if err := json.Unmarshal(data, aliasValue); err != nil {
		return err
	}

	url, err := url2.Parse(aliasValue.RawURL)
	if err != nil {
		return err
	}

	im.ExpandURL = *url
	return nil
}

type OutMessage struct {
	ShortURL string `json:"result"`
}

//Добавьте в сервер новый эндпоинт POST /api/shorten,
//принимающий в теле запроса JSON-объект {"url":"<some_url>"}  и
//возвращающий в ответ объект {"result":"<shorten_url>"}.
func (h *Handler) postHandlerJSON(w http.ResponseWriter, r *http.Request) {
	inData := InMessage{}
	err := json.NewDecoder(r.Body).Decode(&inData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//отображаем переданный урл в уникальный идентификатор
	short, _ := h.shortenURL(&inData.ExpandURL)

	//добавлеям в мапу
	h.storage.Insert(short.Path, inData.ExpandURL.String())

	//проставляем заголовки
	//TODO вынести строковые литералы в константы
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	//пишем тело ответа
	outData := OutMessage{ShortURL: short.String()}
	_ = json.NewEncoder(w).Encode(outData)
}

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
