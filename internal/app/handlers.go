package app

import (
	"bytes"
	"io"
	"net/http"
	url2 "net/url"
)

//Эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор сокращённого URL и
//возвращает ответ с кодом 307 и оригинальным URL в HTTP-заголовке Location.
func (h *Handler) getHandler(w http.ResponseWriter, r *http.Request) {
	//ищем в хранилище соответсвующий полный юрл
	expandURL, ok := h.storage[string(bytes.TrimPrefix([]byte(r.URL.Path), []byte("/")))]
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
	//TODO storage.insert(...)
	short := shortenURL(url)
	//мапим только путь потому что префиксы у всех урлов одинаковые
	h.storage[short.Path] = url.String()

	//устанавливаем статус ответа
	w.WriteHeader(http.StatusCreated)
	//пишем в тело ответа сокращенный url
	w.Write([]byte(short.String()))
}
