package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	url2 "net/url"
)

//Эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор сокращённого URL и
//возвращает ответ с кодом 307 и оригинальным URL в HTTP-заголовке Location.
func (h *Handler) getHandler(w http.ResponseWriter, r *http.Request) {
	//ищем в хранилище соответсвующий полный юрл
	h.lock.RLock()
	expandURL, ok := h.storage[string(bytes.TrimPrefix([]byte(r.URL.Path), []byte("/")))]
	h.lock.RUnlock()

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
	//TODO storage.insert(...)
	short := shortenURL(url)

	h.lock.Lock()
	h.storage[short.Path] = string(b)
	h.lock.Unlock()

	//устанавливаем статус ответа
	w.WriteHeader(http.StatusCreated)
	//пишем в тело ответа сокращенный url
	w.Write([]byte(short.String()))
}

type InMessage struct {
	ExpandUrl url2.URL `json:"url"`
}

//рабоче-крестьянским методом валидируем урл при чтении
func (im *InMessage) UnmarshalJSON(data []byte) error {
	aliasValue := &struct {
		RawUrl string `json:"url"`
	}{}
	if err := json.Unmarshal(data, aliasValue); err != nil {
		return err
	}

	url, err := url2.Parse(aliasValue.RawUrl)
	if err != nil {
		return err
	}

	im.ExpandUrl = *url
	return nil
}

type OutMessage struct {
	ShortUrl string `json:"result"`
}

//Добавьте в сервер новый эндпоинт POST /api/shorten,
//принимающий в теле запроса JSON-объект {"url":"<some_url>"}  и
//возвращающий в ответ объект {"result":"<shorten_url>"}.
func (h *Handler) postJsonHandler(w http.ResponseWriter, r *http.Request) {
	inData := InMessage{}
	err := json.NewDecoder(r.Body).Decode(&inData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//отображаем переданный урл в уникальный идентификатор
	short := shortenURL(&inData.ExpandUrl)
	//добавлеям в мапу
	//TODO storage insert
	h.lock.Lock()
	h.storage[short.Path] = inData.ExpandUrl.String()
	h.lock.Unlock()
	//проставляем заголовки
	//TODO вынести строковые литералы в константы
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	//пишем тело ответа
	outData := OutMessage{ShortUrl: short.String()}
	_ = json.NewEncoder(w).Encode(outData)
}
