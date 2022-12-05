package app

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHandler(t *testing.T) {
	type want struct {
		expandURL    string
		statusCode   int
		expectErr    bool
		maybeErrBody string
	}

	tests := []struct {
		storage map[string]string
		name    string
		request string
		want    want
	}{
		{
			storage: map[string]string{"id1": "http://ya.ru"},
			name:    "simple test 1",
			request: "/id1",
			want: want{
				expandURL:  "http://ya.ru",
				statusCode: http.StatusTemporaryRedirect,
				expectErr:  false,
			},
		},
		{
			storage: map[string]string{"id1": "http://ya.ru"},
			name:    "simple test 2",
			request: "/id2",
			want: want{
				statusCode:   http.StatusBadRequest,
				expectErr:    true,
				maybeErrBody: "Passed short url not found\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler()

			for k, v := range tt.storage {
				h.storage.Insert(k, v, "some_user")
			}

			request := httptest.NewRequest(http.MethodGet, tt.request, nil)

			w := httptest.NewRecorder()
			h.ServeHTTP(w, request)
			result := w.Result()

			b, err := io.ReadAll(result.Body)
			assert.NoError(t, err)

			err = result.Body.Close()
			assert.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			if !tt.want.expectErr {
				assert.Equal(t, tt.want.expandURL, result.Header.Get("Location"))
			} else {
				assert.Equal(t, tt.want.maybeErrBody, string(b))
			}
		})
	}
}

func TestPostHandler(t *testing.T) {
	type want struct {
		statusCode int
		expectErr  bool
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "simple test 1",
			request: "/",
			want: want{
				expectErr:  false,
				statusCode: http.StatusCreated,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler()

			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, request)
			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			actualBody, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			if !tt.want.expectErr {
				assert.True(t, strings.HasPrefix(string(actualBody), h.cfg.BaseURL))
				actualBody = bytes.TrimPrefix(actualBody, []byte(h.cfg.BaseURL))
				respURL, err := url.Parse(string(actualBody))
				require.NoError(t, err)
				_, ok := h.storage.LookUp(respURL.Path)
				assert.True(t, ok)
			}
		})
	}
}

func TestPostJsonHandler(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		expectErr   bool
	}
	tests := []struct {
		name    string
		request string
		body    string
		want    want
	}{
		{
			name:    "simple test 1",
			request: "/api/shorten",
			body:    "{\"url\" :\"http://yandex.ru\"}",
			want: want{
				expectErr:   false,
				statusCode:  http.StatusCreated,
				contentType: "application/json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler()

			request := httptest.NewRequest(http.MethodPost, tt.request, strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			h.ServeHTTP(w, request)
			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			actualBody := OutMessage{}
			err := json.NewDecoder(result.Body).Decode(&actualBody)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			if !tt.want.expectErr {
				assert.True(t, strings.HasPrefix(actualBody.ShortURL, h.cfg.BaseURL))
				actualBody.ShortURL = string(bytes.TrimPrefix(
					[]byte(actualBody.ShortURL),
					[]byte(h.cfg.BaseURL),
				))
				respURL, err := url.Parse(actualBody.ShortURL)
				require.NoError(t, err)
				_, ok := h.storage.LookUp(respURL.Path)
				assert.True(t, ok)
			}
		})
	}
}

func TestDefaultHandler(t *testing.T) {
	type want struct {
		statusCode int
	}
	//this handler can handle all methods except GET and POST
	tests := []struct {
		name    string
		request string
		method  string
		want    want
	}{
		{
			method:  http.MethodDelete,
			name:    "simple test 1",
			request: "/",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			method:  http.MethodConnect,
			name:    "simple test 2",
			request: "/",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler()

			request := httptest.NewRequest(tt.method, tt.request, nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, request)
			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			err := result.Body.Close()
			require.NoError(t, err)
		})
	}
}
