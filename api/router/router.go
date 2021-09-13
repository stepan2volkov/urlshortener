package router

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/stepan2volkov/urlshortener/api/openapi"
	"github.com/stepan2volkov/urlshortener/app"
)

type Router struct {
	http.Handler
	app  *app.App
	host string
}

// NewRouter creates router
func NewRouter(app *app.App, host string) *Router {
	r := chi.NewRouter()
	rt := &Router{app: app}
	r.Use(middleware.Logger)

	// Not the part of main API and can be removed (i.e. after creating frontend)
	r.Get("/", rt.GetMainPage)
	r.Get("/openapi", rt.GetOpenAPI)
	fileServer := http.FileServer(http.Dir("./web/static"))
	r.Get("/static/{filename}", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		http.StripPrefix("/static", fileServer).ServeHTTP(w, r)
	})

	// Main API
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

	_, err := url.ParseRequestURI(requestURL.OriginalURL)
	if err != nil {
		log.Println(err)
		http.Error(w, "url is invalid", http.StatusBadRequest)
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
	w.Header().Add("Content-type", "application/json")
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
	w.Header().Add("Content-type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (rt *Router) GetMainPage(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles("./web/templates/index.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	tmplData := struct{ Host string }{Host: rt.host}
	if err != ts.Execute(w, tmplData) {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}
}

func (rt *Router) GetOpenAPI(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles("./web/templates/openapi.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	tmplData := struct{ Host string }{Host: rt.host}
	if err != ts.Execute(w, tmplData) {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}
}
