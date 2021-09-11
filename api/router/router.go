package router

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/stepan2volkov/urlshortener/api/openapi"
	"github.com/stepan2volkov/urlshortener/app"
)

type Router struct {
	http.Handler
	app *app.App
}

func NewRouter(app *app.App) *Router {
	r := chi.NewRouter()
	rt := &Router{app: app}
	r.Use(middleware.Logger)

	// Not the part of main API and can be removed after creating front-end
	r.Get("/", rt.GetMainPage)
	r.Mount("/", openapi.Handler(rt))

	swagger, err := openapi.GetSwagger()
	if err != nil {
		log.Fatalf("Swagger error: %v\n", err)
	}
	r.Get("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(swagger)
	})

	rt.Handler = r
	return rt
}

type RequestURL struct {
	OriginalURL string `json:"originalURL"`
}
type ResponseURL struct {
	ShortURL string `json:"shortURL"`
	StatsURL string `json:"statsURL"`
}

type Stats struct {
	ShortURL     string `json:"shortURL"`
	NumRedirects int    `json:"numRedirects"`
}

func (rt *Router) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	requestURL := &RequestURL{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(requestURL); err != nil {
		log.Println(err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	url, err := rt.app.CreateURL(r.Context(), requestURL.OriginalURL)
	if err != nil {
		log.Println(err)
		http.Error(w, "couldn't create short url", http.StatusInternalServerError)
		return
	}

	responseURL := &ResponseURL{
		ShortURL: url.ShortURL,
		StatsURL: "stats/" + url.ShortURL,
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(responseURL)
}

func (rt *Router) RedirectURL(w http.ResponseWriter, r *http.Request, shortURL string) {
	url, err := rt.app.GetRedirectURL(r.Context(), shortURL)
	if err != nil {
		log.Println(err)
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, url.OriginalURL, http.StatusSeeOther)
}

func (rt *Router) GetStats(w http.ResponseWriter, r *http.Request, shortURL string) {
	stats, err := rt.app.GetStats(r.Context(), shortURL)
	if err != nil {
		log.Println(err)
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	response := &Stats{
		ShortURL:     stats.ShortURL,
		NumRedirects: stats.NumRedirects,
	}
	_ = json.NewEncoder(w).Encode(response)
}

func (rt *Router) GetMainPage(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles("./web/templates/index.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	if err != ts.Execute(w, nil) {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}
}
