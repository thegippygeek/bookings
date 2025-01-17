package main

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/thegippygeek/bookings/pkg/config"
	"github.com/thegippygeek/bookings/pkg/handlers"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)
	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)

	fileserver := http.FileServer(http.Dir("./static/"))

	mux.Handle("/static/*", http.StripPrefix("/static", fileserver))

	return mux
}
