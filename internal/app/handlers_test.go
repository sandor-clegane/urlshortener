package app

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAPIServer_getHandler(t *testing.T) {
	type want struct {
		expandURL  string
		statusCode int
		expectErr  bool
	}
	//TODO refactor storage to interface
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
				expandURL:  "",
				statusCode: http.StatusBadRequest,
				expectErr:  true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler()
			h.storage = tt.storage

			request := httptest.NewRequest(http.MethodGet, tt.request, nil)

			w := httptest.NewRecorder()
			h.ServeHTTP(w, request)
			result := w.Result()

			err := result.Body.Close()
			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			if !tt.want.expectErr {
				assert.Equal(t, tt.want.expandURL, result.Header.Get("Location"))
			}
		})
	}
}

func TestAPIServer_postHandler(t *testing.T) {
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
				strings.Contains(string(actualBody), "http://localhost:8080/")
				actualBody = bytes.TrimPrefix(actualBody, []byte("http://localhost:8080/"))
				respURL, err := url.Parse(string(actualBody))
				require.NoError(t, err)
				assert.Contains(t, h.storage, respURL.Path)
			}
		})
	}
}

func TestAPIServer_defaultHandler(t *testing.T) {
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
