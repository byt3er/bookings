package main

import (
	"net/http"

	"github.com/byt3er/bookings/internals/helpers"
	"github.com/justinas/nosurf"
)

// NoSurf is the csrf protection middleware
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

// SessionLoad loads and saves session data for current request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

// What this is doing,
// this is the exactly the same logic as the noSurf() or as sessionLoad()
// but it's our own custom middleware that actually has access to the request
// so that be can call helpers.IsAuthenticate()
// we use this middleware to protect routes in our routes file
// So this will make sure that only people who are logged in actually have
// access to the routes that we want to protect
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !helpers.IsAuthenticate(r) {
			session.Put(r.Context(), "error", "Login First!")
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		// if this doesn't fail just pass onto the next middleware
		// and the request lifecycle continues on its way
		next.ServeHTTP(w, r)

	})
}
