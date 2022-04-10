package router

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/stepan2volkov/urlshortener/api/openapi"
	"github.com/stepan2volkov/urlshortener/app"
)

const (
	namespace    = "urlshortener"
	labelMethod  = "method"
	labelHandler = "handler"
	labelStatus  = "status"
)

type Router struct {
	http.Handler
	app              *app.App
	latencyHistogram *prometheus.HistogramVec
	logger           *zap.Logger
}

// NewRouter creates router
func NewRouter(app *app.App, logger *zap.Logger) *Router {
	r := chi.NewRouter()
	rt := &Router{app: app}
	rt.logger = logger
	if err := rt.init(); err != nil {
		logger.Fatal("error when initializing metrics",
			zap.Error(err))
	}

	r.Use(middleware.Logger)

	// Not the part of main API and can be removed (i.e. after creating frontend)
	r.Get("/", rt.GetMainPage)
	r.Get("/openapi", rt.GetOpenAPI)
	fileServer := http.FileServer(http.Dir("./web/static"))
	r.Get("/static/{filename}", func(w http.ResponseWriter, r *http.Request) {
		rt.logger.Debug("file was requested", zap.String("url", r.RequestURI))
		http.StripPrefix("/static", fileServer).ServeHTTP(w, r)
	})

	// Main API
	r.Mount("/", openapi.Handler(rt))

	swagger, err := openapi.GetSwagger()
	if err != nil {
		logger.Fatal("error when getting swagger",
			zap.Error(err))
	}
	r.Get("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(swagger)
	})
	r.Get("/metrics", promhttp.Handler().ServeHTTP)

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

func (rt *Router) init() error {
	rt.latencyHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "latency",
		Buckets: []float64{0, 25, 50, 75, 100, 200, 400, 600, 800, 1000, 2000,
			4000, 6000},
	}, []string{labelMethod, labelHandler, labelStatus})

	return prometheus.Register(rt.latencyHistogram)
}

func (rt *Router) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	startedAt := time.Now()

	requestURL := &RequestURL{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(requestURL); err != nil {
		rt.logger.Info("error when encoding request body",
			zap.Error(err))
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	_, err := url.ParseRequestURI(requestURL.OriginalURL)
	if err != nil {
		rt.logger.Info("error when parse url",
			zap.Error(err))
		rt.latencyHistogram.With(prometheus.Labels{
			labelMethod:  http.MethodPost,
			labelStatus:  strconv.Itoa(http.StatusBadRequest),
			labelHandler: "create_short_url",
		}).Observe(sinceInMilliseconds(startedAt))
		http.Error(w, "url is invalid", http.StatusBadRequest)
		return
	}

	url, err := rt.app.CreateURL(r.Context(), requestURL.OriginalURL)
	if err != nil {
		rt.logger.Error("error when creating url",
			zap.Error(err))
		rt.latencyHistogram.With(prometheus.Labels{
			labelMethod:  http.MethodPost,
			labelStatus:  strconv.Itoa(http.StatusInternalServerError),
			labelHandler: "create_short_url",
		}).Observe(sinceInMilliseconds(startedAt))

		http.Error(w, "couldn't create short url", http.StatusInternalServerError)
		return
	}

	responseURL := &ResponseURL{
		ShortURL: "/" + url.ShortURL,
		StatsURL: "/stats/" + url.ShortURL,
	}
	rt.latencyHistogram.With(prometheus.Labels{
		labelMethod:  http.MethodPost,
		labelStatus:  strconv.Itoa(http.StatusCreated),
		labelHandler: "create_short_url",
	}).Observe(sinceInMilliseconds(startedAt))

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(responseURL)
}

func (rt *Router) RedirectURL(w http.ResponseWriter, r *http.Request, shortURL string) {
	startedAt := time.Now()

	url, err := rt.app.GetRedirectURL(r.Context(), shortURL)
	if err != nil {
		rt.logger.Info("error when getting redirect url",
			zap.Error(err),
			zap.String("short_url", shortURL))
		http.Error(w, "not found", http.StatusNotFound)

		rt.latencyHistogram.With(prometheus.Labels{
			labelMethod:  http.MethodGet,
			labelStatus:  strconv.Itoa(http.StatusNotFound),
			labelHandler: "redirect_url",
		}).Observe(sinceInMilliseconds(startedAt))

		return
	}
	rt.latencyHistogram.With(prometheus.Labels{
		labelMethod:  http.MethodGet,
		labelStatus:  strconv.Itoa(http.StatusSeeOther),
		labelHandler: "redirect_url",
	}).Observe(sinceInMilliseconds(startedAt))

	http.Redirect(w, r, url.OriginalURL, http.StatusSeeOther)
}

func (rt *Router) GetStats(w http.ResponseWriter, r *http.Request, shortURL string) {
	startedAt := time.Now()

	stats, err := rt.app.GetStats(r.Context(), shortURL)
	if err != nil {
		rt.logger.Error("error when getting stas",
			zap.Error(err),
			zap.String("short_url", shortURL))
		rt.latencyHistogram.With(prometheus.Labels{
			labelMethod:  http.MethodGet,
			labelStatus:  strconv.Itoa(http.StatusNotFound),
			labelHandler: "get_stats",
		}).Observe(sinceInMilliseconds(startedAt))

		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	response := &Stats{
		ShortURL:     stats.ShortURL,
		NumRedirects: stats.NumRedirects,
	}
	rt.latencyHistogram.With(prometheus.Labels{
		labelMethod:  http.MethodGet,
		labelStatus:  strconv.Itoa(http.StatusOK),
		labelHandler: "get_stats",
	}).Observe(sinceInMilliseconds(startedAt))

	w.Header().Add("Content-type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (rt *Router) GetMainPage(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles("./web/templates/index.html")
	if err != nil {
		rt.logger.Warn("error when getting main page",
			zap.Error(err))
		http.Error(w, "Internal Server Error", 500)

		return
	}

	if err != ts.Execute(w, nil) {
		rt.logger.Warn("error when processing template for main-page",
			zap.Error(err))
		http.Error(w, "Internal Server Error", 500)
		return
	}
}

func (rt *Router) GetOpenAPI(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles("./web/templates/openapi.html")
	if err != nil {
		rt.logger.Warn("error when getting swagger page",
			zap.Error(err))
		http.Error(w, "Internal Server Error", 500)
		return
	}

	if err != ts.Execute(w, nil) {
		rt.logger.Warn("error when processing template for swagger",
			zap.Error(err))
		http.Error(w, "Internal Server Error", 500)
		return
	}
}

func sinceInMilliseconds(startTime time.Time) float64 {
	return float64(time.Since(startTime).Nanoseconds()) / 1e6
}
