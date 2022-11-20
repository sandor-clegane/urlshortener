package app

import (
	"encoding/json"
	"io"
	"net/http"
	url2 "net/url"
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
