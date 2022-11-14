package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// something that hold post

type postData struct {
	key   string
	value string
}

// variable for actual test
// slice of struct
var theTests = []struct {
	name   string     // name of the test; any given name
	url    string     // the  path which match by our route
	method string     // the method ; is it GET or POST
	params []postData // things that are being posted
	// something to say whether or not a test has passed
	// what kind of response code are we getting back from the web server
	// if its 200, everything worked the way it's supposed to.
	// if its 404, page not found
	// if its 300 range, its a redirect
	expectedStatusCode int
}{
	{"home", "/", "GET", []postData{}, http.StatusOK},
	{"about", "/about", "GET", []postData{}, http.StatusOK},
	{"gq", "/generals-quaters", "GET", []postData{}, http.StatusOK},
	{"ms", "/majors-suite", "GET", []postData{}, http.StatusOK},
	{"sa", "/search-availability", "GET", []postData{}, http.StatusOK},
	{"contact", "/contact", "GET", []postData{}, http.StatusOK},
	{"mr", "/make-reservation", "GET", []postData{}, http.StatusOK},
	{"post-search-avail", "/search-availability", "POST", []postData{
		{key: "start", value: "2020-01-01"},
		{key: "end", value: "2020-01-02"},
	}, http.StatusOK},
	{"post-search-avail-json", "/search-availability-json", "POST", []postData{
		{key: "start", value: "2020-01-01"},
		{key: "end", value: "2020-01-02"},
	}, http.StatusOK},
	{"post-make-reservation", "/make-reservation", "POST", []postData{
		{key: "first_name", value: "Jhon"},
		{key: "last_name", value: "Smith"},
		{key: "email", value: "me@here.com"},
		{key: "phone", value: "555-555-5555"},
	}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	// test server
	// once the server is created it's going to fire up
	// for the life of the test
	// its  going to listen on port..
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	//table test
	for _, e := range theTests {
		if e.method == "GET" {
			// make a request a client as we were a
			// web browser accessing a web page
			resp, err := ts.Client().Get(ts.URL + e.url)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}
			//check our response
			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("for %s, expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)

			}
		} else {
			values := url.Values{} //built in type that holds information as a post request vairable
			for _, x := range e.params {
				values.Add(x.key, x.value)
			}
			resp, err := ts.Client().PostForm(ts.URL+e.url, values)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}
			//check our response
			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("for %s, expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)

			}

		}
	}
}
