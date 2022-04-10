package router

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/stepan2volkov/urlshortener/app"
	"github.com/stepan2volkov/urlshortener/db/memstore"
	"go.uber.org/zap"
)

func TestRouter_CreateShortURL(t *testing.T) {
	tests := []struct {
		name    string
		request string
		code    int
	}{
		{name: "201", request: `{"originalURL": "https://google.com"}`, code: 201},
		{name: "400", request: `{"originalURL": ";DROP TABLE urls"}`, code: 400},
	}

	store := memstore.NewMemStore()
	logger := zap.NewNop()
	app := app.NewApp(store, logger, opentracing.NoopTracer{})
	router := NewRouter(app, logger, opentracing.NoopTracer{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/", strings.NewReader(tt.request))
			router.CreateShortURL(w, r)
			if w.Code != tt.code {
				t.Errorf("Unexpected status code: want - %v, got %v\n", tt.code, w.Code)
			}
		})
	}
}

func TestRouter_RedirectURL(t *testing.T) {
	tests := []struct {
		name        string
		originalURL string
		shortURL    string
		code        int
	}{
		{name: "303", originalURL: "https://google.com", code: 303},
		{name: "404", originalURL: "", shortURL: "999", code: 404},
	}

	store := memstore.NewMemStore()
	logger := zap.NewNop()
	app := app.NewApp(store, logger, opentracing.NoopTracer{})
	router := NewRouter(app, logger, opentracing.NoopTracer{})

	for i, tt := range tests {
		if tt.originalURL != "" {
			url, err := app.CreateURL(context.Background(), tt.originalURL)
			if err != nil {
				t.Errorf("error when create url \"%v\": %v\n", tt.originalURL, err)
			}
			tests[i].shortURL = url.ShortURL
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/"+tt.shortURL, nil)
			router.RedirectURL(w, r, tt.shortURL)
			if w.Code != tt.code {
				t.Errorf("Unexpected status code: want - %v, got %v\n", tt.code, w.Code)
			}
		})
	}
}

func TestRouter_GetStats(t *testing.T) {
	tests := []struct {
		name         string
		originalURL  string
		shortURL     string
		code         int
		numRedirects int
	}{
		{name: "200", originalURL: "https://golang.org", code: 200, numRedirects: 0},
		{name: "200-13", originalURL: "https://google.com", code: 200, numRedirects: 13},
		{name: "200-500", originalURL: "https://google.com/search?q=golang", code: 200, numRedirects: 500},
		{name: "404", originalURL: "", shortURL: "999", code: 404},
	}

	store := memstore.NewMemStore()
	logger := zap.NewNop()
	app := app.NewApp(store, logger, opentracing.NoopTracer{})
	router := NewRouter(app, logger, opentracing.NoopTracer{})

	for i, tt := range tests {
		if tt.originalURL != "" {
			url, err := app.CreateURL(context.Background(), tt.originalURL)
			if err != nil {
				t.Errorf("error when create url \"%v\": %v\n", tt.originalURL, err)
			}
			tests[i].shortURL = url.ShortURL
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.numRedirects; i++ {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/"+tt.shortURL, nil)
				router.RedirectURL(w, r, tt.shortURL)
			}
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/stats/"+tt.shortURL, nil)
			router.GetStats(w, r, tt.shortURL)

			if w.Code != tt.code {
				t.Errorf("Unexpected status code: want - %v, got %v\n", tt.code, w.Code)
			}
			if tt.code == 200 {
				stats := &Stats{}
				err := json.NewDecoder(w.Body).Decode(stats)
				if err != nil {
					t.Errorf("[%v] Error when decode: %v\n", tt.name, err)
				}
				if stats.NumRedirects != tt.numRedirects {
					t.Errorf("[%v] Unexpected redirect num: want - %v, got - %v\n", tt.name, tt.numRedirects, stats.NumRedirects)
				}
			}
		})
	}
}
