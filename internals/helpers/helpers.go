// put things that are useful that can use in various parts of my application
package helpers

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/byt3er/bookings/internals/config"
)

var app *config.AppConfig

func NewHelpers(a *config.AppConfig) {
	app = a
}

// errors made by the client
func ClientError(w http.ResponseWriter, status int) {
	// write to the infolog
	app.InfoLog.Println("Client error with status of ", status)
	http.Error(w, http.StatusText(status), status)
}

// something went wrong with server
func ServerError(w http.ResponseWriter, err error) {
	// get a trace of the error; nature of the error
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	// When we go into production we can have our log write to a file
	// or might send a text or send an email
	// It can let someone know that something went wrong and then they
	// can open the log file and look at the error message and go fix the problem
	app.ErrorLog.Println(trace)
	// feedback for the client
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
