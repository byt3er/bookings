package main

import (
	"net/http"

	"github.com/manoj-kr-borah/bookings/internals/config"
	"github.com/manoj-kr-borah/bookings/internals/handlers"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)
	mux.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir("./static/"))))
	mux.Handle("/favicon.ico", http.NotFoundHandler())
	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/generals-quaters", handlers.Repo.Generals)
	mux.Get("/majors-suite", handlers.Repo.Majors)
	mux.Get("/contact", handlers.Repo.Contact)
	mux.Get("/make-reservation", handlers.Repo.Reservation)

	mux.Get("/search-availability", handlers.Repo.Availability)
	mux.Post("/search-availability", handlers.Repo.PostAvailability)
	mux.Post("/search-availability-json", handlers.Repo.AvailabilityJSON)

	return mux
}
