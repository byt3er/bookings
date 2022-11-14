package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestNoSurf(t *testing.T) {
	var myH myHandler

	//NoSurf takes a *http.Handler and returns *http.Handler
	h := NoSurf(&myH)

	//Check whether h is  actually a handler
	switch v := h.(type) { //v stores the type of h
	case http.Handler:
		// do nothing , passed the test
	default:
		t.Error(fmt.Sprintf(" type is not http.Handler but is %T", v))

	}

}
func TestSessionLoad(t *testing.T) {
	var myH myHandler

	//NoSurf takes a *http.Handler and returns *http.Handler
	h := SessionLoad(&myH)

	//Check whether h is  actually a handler
	switch v := h.(type) { //v stores the type of h
	case http.Handler:
		// do nothing , passed the test
	default:
		t.Error(fmt.Sprintf(" type is not http.Handler but is %T", v))

	}

}
