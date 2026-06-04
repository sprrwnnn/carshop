package httpserver

import (
	"carshop/internal/config"
	"net/http"
	"net/http/pprof"

	"carshop/internal/application/http_server/handlers"
	"carshop/internal/application/http_server/middleware"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Router struct {
	router chi.Router
	env    config.Env

	handler *handlers.Handler

	loggingMiddleware middleware.Middleware
	metricsMiddleware middleware.Middleware
}

func NewRouter(
	env config.Env,
	handler *handlers.Handler,

	loggingMiddleware middleware.Middleware,
	metricsMiddleware middleware.Middleware,
) (*Router, error) {
	return &Router{
		router: chi.NewRouter(),
		env:    env,

		handler: handler,

		loggingMiddleware: loggingMiddleware,
		metricsMiddleware: metricsMiddleware,
	}, nil
}

func (r *Router) Router() chi.Router {
	return r.router
}

func (r *Router) SetupV1() {
	r.router.Use(r.loggingMiddleware.Intercept)
	r.router.Use(r.metricsMiddleware.Intercept)

	r.router.Route("/api/v1", func(rou chi.Router) {
		if r.env != config.EnvProduction {
			rou.Mount("/healthcheck", healthcheckRouter(r.handler))
		}

		rou.Mount("/cars", r.CarsRouter())
		rou.Mount("/quirky", r.QuirkyRouter())
		rou.Mount("/metrics", promhttp.Handler())
	})

	// --------------- SYSTEM INFO ------------
	if r.env != config.EnvProduction {
		r.router.Mount("/debug", systemInfoRouter())
	}

	// --------------- SWAGGER UI ------------
	r.router.Get("/swagger", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/swagger/index.html", http.StatusMovedPermanently)
	})
	r.router.Get("/swagger/*", httpSwagger.WrapHandler)
}

func healthcheckRouter(h *handlers.Handler) chi.Router {
	rou := chi.NewRouter()

	rou.Get("/", h.GetBuildInfoHandler)

	return rou
}

// System info handlers
func systemInfoRouter() chi.Router {
	rou := chi.NewRouter()

	rou.Route("/pprof", func(rou chi.Router) {
		rou.Get("/", pprof.Index)
		rou.Get("/cmdline", pprof.Cmdline)
		rou.Get("/profile", pprof.Profile)
		rou.Get("/symbol", pprof.Symbol)
		rou.Get("/trace", pprof.Trace)
		rou.Mount("/goroutine", pprof.Handler("goroutine"))
		rou.Mount("/heap", pprof.Handler("heap"))
		rou.Mount("/threadcreate", pprof.Handler("threadcreate"))
		rou.Mount("/block", pprof.Handler("block"))
	})

	return rou
}

// Cars handlers
func (r *Router) CarsRouter() chi.Router {
	rou := chi.NewRouter()

	rou.Route("/q", func(rou chi.Router) {
		rou.Get("/", r.handler.GetCarsHandler)
		rou.Get("/{id:[0-9]+}", r.handler.GetCarByIDHandler)
	})

	rou.Route("/c", func(rou chi.Router) {
		rou.Post("/", r.handler.CreateCarHandler)
		rou.Patch("/{id:[0-9]+}", r.handler.UpdateCarHandler)
		rou.Delete("/{id:[0-9]+}", r.handler.DeleteCarHandler)
	})

	return rou
}

// Quirky handlers
func (r *Router) QuirkyRouter() chi.Router {
	rou := chi.NewRouter()

	rou.Get("/panic", r.handler.PanicHandler)
	rou.Get("/slow", r.handler.SlowHandler)

	return rou
}
