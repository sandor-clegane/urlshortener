package url

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

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
	expandURL, err := h.us.ExpandURL(r.Context(), r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Add("Location", expandURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

//ShortenURL эндпоинт POST / принимает в теле запроса строку URL для сокращения
//и возвращает ответ с кодом 201 и сокращённым URL в виде текстовой строки в теле.
func (h *URLhandlerImpl) ShortenURL(w http.ResponseWriter, r *http.Request) {
	authCookie, err := r.Cookie("userID")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID, err := h.cs.ExtractValue(authCookie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rawurl, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	short, err := h.us.ShortenURL(r.Context(), userID, string(rawurl))
	if err != nil {
		var violationError *myerrors.UniqueViolation
		if errors.As(err, &violationError) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(violationError.ExistedShortURL))
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(short))
}

//ShortenURLwJSON Добавьте в сервер новый эндпоинт POST /api/shorten,
//принимающий в теле запроса JSON-объект {"url":"<some_url>"}  и
//возвращающий в ответ объект {"result":"<shorten_url>"}.
func (h *URLhandlerImpl) ShortenURLwJSON(w http.ResponseWriter, r *http.Request) {
	authCookie, err := r.Cookie("userID")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID, err := h.cs.ExtractValue(authCookie)
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
	if err != nil {
		var violationError *myerrors.UniqueViolation
		if errors.As(err, &violationError) {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode([]byte(violationError.ExistedShortURL))
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	outData := common.OutMessage{ShortURL: short}
	json.NewEncoder(w).Encode(outData)
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
	authCookie, err := r.Cookie("userID")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID, err := h.cs.ExtractValue(authCookie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	listOfURL, err := h.us.GetAllURL(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(listOfURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (h *URLhandlerImpl) ShortenSomeURL(w http.ResponseWriter, r *http.Request) {
	authCookie, err := r.Cookie("userID")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID, err := h.cs.ExtractValue(authCookie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var expandURLwIDslice []common.PairURLwithCIDin
	err = json.NewDecoder(r.Body).Decode(&expandURLwIDslice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	shortURLwIDslice, err := h.us.ReduceSeveralURL(r.Context(), userID, expandURLwIDslice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(shortURLwIDslice)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}
