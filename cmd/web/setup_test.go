package main

import (
	"net/http"
	"os"
	"testing"
)

// before you start running the test
// do something inside this function
// then run the test and exit
func TestMain(m *testing.M) {

	os.Exit(m.Run())
}

// Variables
type myHandler struct{}

func (m *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
