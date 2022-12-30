package url

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	_ "github.com/lib/pq"
	"github.com/sandor-clegane/urlshortener/internal/common"
	"github.com/sandor-clegane/urlshortener/internal/common/myerrors"
	"github.com/sandor-clegane/urlshortener/internal/config"
	"github.com/sandor-clegane/urlshortener/internal/service/cookie"
	"github.com/sandor-clegane/urlshortener/internal/service/shortener"
	"github.com/sandor-clegane/urlshortener/internal/storages"
)

type URLhandlerImpl struct {
	us shortener.URLshortenerService
	cs cookie.CookieService
}

func New(stg storages.Storage, cfg config.Config) URLHandler {
	return &URLhandlerImpl{
		cs: cookie.New(cfg.Key),
		us: shortener.New(stg, cfg.BaseURL),
	}
}

func (h *URLhandlerImpl) GetAuthorizationMiddleware() func(next http.Handler) http.Handler {
	return h.cs.Authentication
}

//ExpandURL Эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор сокращённого URL и
//возвращает ответ с кодом 307 и оригинальным URL в HTTP-заголовке Location.
func (h *URLhandlerImpl) ExpandURL(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")
	if shortURL == "" {
		http.Error(w, "No id param in URL", http.StatusBadRequest)
		return
	}
	expandURL, err := h.us.ExpandURL(r.Context(), shortURL)
	if err != nil {
		var deletedErr *myerrors.DeleteViolation
		if errors.As(err, &deletedErr) {
			w.WriteHeader(http.StatusGone)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	w.Header().Add("Location", expandURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

//ShortenURL эндпоинт POST / принимает в теле запроса строку URL для сокращения
//и возвращает ответ с кодом 201 и сокращённым URL в виде текстовой строки в теле.
func (h *URLhandlerImpl) ShortenURL(w http.ResponseWriter, r *http.Request) {
	userID, err := h.cs.GetUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rawURL, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	short, err := h.us.ShortenURL(r.Context(), userID, string(rawURL))
	if err == nil {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(short))
		return
	}
	var violationError *myerrors.UniqueViolation
	if errors.As(err, &violationError) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(violationError.ExistedShortURL))
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

//ShortenURLwJSON Добавьте в сервер новый эндпоинт POST /api/shorten,
//принимающий в теле запроса JSON-объект {"url":"<some_url>"}  и
//возвращающий в ответ объект {"result":"<shorten_url>"}.
func (h *URLhandlerImpl) ShortenURLwJSON(w http.ResponseWriter, r *http.Request) {
	userID, err := h.cs.GetUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	inData := common.InMessage{}
	err = json.NewDecoder(r.Body).Decode(&inData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	short, err := h.us.ShortenURL(r.Context(), userID, inData.ExpandURL.String())
	if err == nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		outData := common.OutMessage{ShortURL: short}
		json.NewEncoder(w).Encode(outData)
		return
	}
	var violationError *myerrors.UniqueViolation
	if errors.As(err, &violationError) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(
			common.OutMessage{ShortURL: violationError.ExistedShortURL})
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

//GetAllURL Иметь хендлер GET /api/user/urls, который сможет вернуть
//пользователю все когда-либо сокращённые им URL в формате:
//[
//    {
//        "short_url": "http://...",
//        "original_url": "http://..."
//    },
//    ...
//]
//При отсутствии сокращённых пользователем URL хендлер должен отдавать HTTP-статус 204 No Content.
func (h *URLhandlerImpl) GetAllURL(w http.ResponseWriter, r *http.Request) {
	userID, err := h.cs.GetUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	listOfURL, err := h.us.GetAllURL(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(listOfURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *URLhandlerImpl) ShortenSomeURL(w http.ResponseWriter, r *http.Request) {
	userID, err := h.cs.GetUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var expandURLwIDslice []common.PairURLwithCIDin
	err = json.NewDecoder(r.Body).Decode(&expandURLwIDslice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortURLwIDslice, err := h.us.ShortenSomeURL(r.Context(), userID, expandURLwIDslice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(shortURLwIDslice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

//DeleteMultipleURLs асинхронный хендлер DELETE /api/user/urls, который
//принимает список идентификаторов сокращённых URL для удаления в формате:
// [ "a", "b", "c", "d", ...]
//В случае успешного приёма запроса хендлер должен возвращать HTTP-статус 202 Accepted.
func (h *URLhandlerImpl) DeleteMultipleURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := h.cs.GetUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var urlIDList []string
	err = json.NewDecoder(r.Body).Decode(&urlIDList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.us.DeleteMultipleURLs(r.Context(), userID, urlIDList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// PingConnectionDB хендлер GET /ping, который при запросе проверяет соединение с базой данных.
//При успешной проверке хендлер должен вернуть HTTP-статус 200 OK,
//при неуспешной — 500 Internal Server Error
func (h *URLhandlerImpl) PingConnectionDB(w http.ResponseWriter, r *http.Request) {
	err := h.us.Ping(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
