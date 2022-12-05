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
	expandURL, ok := h.storage.LookUp(r.URL.Path)
	if !ok {
		http.Error(w, "Passed short url not found", http.StatusBadRequest)
		return
	}
	w.Header().Add("Location", expandURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

//ндпоинт POST / принимает в теле запроса строку URL для сокращения
//и возвращает ответ с кодом 201 и сокращённым URL в виде текстовой строки в теле.
func (h *Handler) postHandler(w http.ResponseWriter, r *http.Request) {
	authCookie, err := r.Cookie("userID")
	if err != nil {
		http.Error(w, "No authentication cookie", http.StatusBadRequest)
		return
	}
	userID, err := h.cookie.extractValue(authCookie)
	if err != nil {
		http.Error(w, "Unauthorized user", http.StatusBadRequest)
		return
	}

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

	h.storage.Insert(short.Path, string(b), userID)
	//устанавливаем статус ответа
	//пишем в тело ответа сокращенный url
	w.WriteHeader(http.StatusCreated)
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
	authCookie, err := r.Cookie("userID")
	if err != nil {
		http.Error(w, "No authentication cookie", http.StatusBadRequest)
		return
	}
	userID, err := h.cookie.extractValue(authCookie)
	if err != nil {
		http.Error(w, "Unauthorized user", http.StatusBadRequest)
		return
	}

	inData := InMessage{}
	err = json.NewDecoder(r.Body).Decode(&inData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//отображаем переданный урл в уникальный идентификатор
	short, _ := h.shortenURL(&inData.ExpandURL)

	//добавлеям в мапу
	h.storage.Insert(short.Path, inData.ExpandURL.String(), userID)

	//проставляем заголовки
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	//пишем тело ответа
	outData := OutMessage{ShortURL: short.String()}
	_ = json.NewEncoder(w).Encode(outData)
}

type PairURL struct {
	ShortURL  string `json:"short_url"`
	ExpandURL string `json:"original_url"`
}

//Иметь хендлер GET /api/user/urls, который сможет вернуть
//пользователю все когда-либо сокращённые им URL в формате:
//[
//    {
//        "short_url": "http://...",
//        "original_url": "http://..."
//    },
//    ...
//]
//При отсутствии сокращённых пользователем URL хендлер должен отдавать HTTP-статус 204 No Content.
func (h *Handler) getAllURLHandler(w http.ResponseWriter, r *http.Request) {
	authCookie, err := r.Cookie("userID")
	if err != nil {
		http.Error(w, "No authentication cookie", http.StatusBadRequest)
		return
	}
	userID, err := h.cookie.extractValue(authCookie)
	if err != nil {
		http.Error(w, "Unauthorized user", http.StatusBadRequest)
		return
	}

	listOfURL, ok := h.storage.GetPairsById(userID)
	if !ok {
		http.Error(w, "User didn`t shorten any URL", http.StatusNoContent)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(listOfURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
