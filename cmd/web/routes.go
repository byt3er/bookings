package main

import (
	"net/http"

	"github.com/byt3er/bookings/internals/config"
	"github.com/byt3er/bookings/internals/handlers"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	// middlewares
	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	// routes for serving static content
	mux.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir("./static/"))))
	mux.Handle("/favicon.ico", http.NotFoundHandler())

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/contact", handlers.Repo.Contact)

	mux.Get("/generals-quaters", handlers.Repo.Generals)
	mux.Get("/majors-suite", handlers.Repo.Majors)

	mux.Get("/search-availability", handlers.Repo.Availability)
	mux.Post("/search-availability", handlers.Repo.PostAvailability)
	mux.Post("/search-availability-json", handlers.Repo.AvailabilityJSON)
	// here {id} is the variable name
	mux.Get("/choose-room/{id}", handlers.Repo.ChooseRoom)
	mux.Get("/book-room", handlers.Repo.BookRoom)

	mux.Get("/make-reservation", handlers.Repo.Reservation)
	mux.Post("/make-reservation", handlers.Repo.PostReservation)
	mux.Get("/reservation-summary", handlers.Repo.ReservationSummary)

	mux.Get("/user/login", handlers.Repo.ShowLogin)
	mux.Post("/user/login", handlers.Repo.PostShowLogin)
	mux.Get("/user/logout", handlers.Repo.Logout)

	// need to set a pattern "/admin"
	// anything that starts with admin will be handled by this function(mux.Route function)
	// on the mux router
	mux.Route("/admin", func(mux chi.Router) {
		// we're going to use Auth middleware to only apply to things
		// that are inside thid mux.Route func
		// mux.Use(Auth)

		mux.Get("/dashboard", handlers.Repo.AdminDashboard)
		mux.Get("/reservations-new", handlers.Repo.AdminNewReservations)
		mux.Get("/reservations-all", handlers.Repo.AdminAllReservations)
		mux.Get("/reservations-calendar", handlers.Repo.AdminReservationsCalender)

		// src and id are matching parameters or matching parts of the route
		mux.Get("/reservations/{src}/{id}", handlers.Repo.AdminShowReservation)
	})
	return mux
}
