package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	url2 "net/url"
)

func (s *APIServer) myHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			s.getHandler(writer, request)
		case http.MethodPost:
			s.postHandler(writer, request)
		default:
			s.defaultHandler(writer, request)
		}
	}
}

//Эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор сокращённого URL и
//возвращает ответ с кодом 307 и оригинальным URL в HTTP-заголовке Location.
func (s *APIServer) getHandler(w http.ResponseWriter, r *http.Request) {
	//ищем в хранилище соответсвующий полный юрл
	expandURL, ok := s.storage[string(bytes.TrimPrefix([]byte(r.URL.Path), []byte("/")))]
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
func (s *APIServer) postHandler(w http.ResponseWriter, r *http.Request) {
	//читаем тело запроса
	b, err := io.ReadAll(r.Body)
	//обрабатываем ошибку
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//тело запроса должно содержать валидный url
	url, err := url2.Parse(string(b))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//сокращаем юрл и добавляем его в хранилище
	//TODO storage.insert(...)
	short := shortenURL(url)
	fmt.Println(short.String())
	s.storage[short.Path] = url.String()

	fmt.Println("Printing map")
	for k, v := range s.storage {
		fmt.Printf("key:%v   value:%v\n", k, v)
	}

	//устанавливаем статус ответа
	w.WriteHeader(http.StatusCreated)
	//пишем в тело ответа сокращенный url
	w.Write([]byte(short.String()))
}

func (s *APIServer) defaultHandler(w http.ResponseWriter, r *http.Request) {
	// этот обработчик принимает все запросы, кроме отправленных методами GET и POST
	http.Error(w, "This method is not allowed", http.StatusMethodNotAllowed)
}
