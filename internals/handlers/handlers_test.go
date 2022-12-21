package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/byt3er/bookings/internals/models"
	"github.com/go-chi/chi"
)

// something that hold post

type postData struct {
	key   string
	value string
}

// variable for actual test
// slice of struct
var theTests = []struct {
	name   string // name of the test; any given name
	url    string // the  path which match by our route
	method string // the method ; is it GET or POST
	// params []postData // things that are being posted
	// something to say whether or not a test has passed
	// what kind of response code are we getting back from the web server
	// if its 200, everything worked the way it's supposed to.
	// if its 404, page not found
	// if its 300 range, its a redirect
	expectedStatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"gq", "/generals-quaters", "GET", http.StatusOK},
	{"ms", "/majors-suite", "GET", http.StatusOK},
	{"sa", "/search-availability", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},
	{"non-existent", "/green/eggs/and/han", "GET", http.StatusNotFound},

	// new routes
	{"login", "/user/login", "GET", http.StatusOK},
	{"logout", "/user/logout", "GET", http.StatusOK},
	{"dashboard", "/admin/dashboard", "GET", http.StatusOK},
	{"new res", "/admin/reservations-new", "GET", http.StatusOK},
	{"all res", "/admin/reservations-all", "GET", http.StatusOK},
	{"show res", "/admin/reservations/new/7/show", "GET", http.StatusOK},
	// {"mr", "/make-reservation", "GET", []postData{}, http.StatusOK},

	// {"post-search-avail", "/search-availability", "POST", []postData{
	// 	{key: "start", value: "2020-01-01"},
	// 	{key: "end", value: "2020-01-02"},
	// }, http.StatusOK},
	// {"post-search-avail-json", "/search-availability-json", "POST", []postData{
	// 	{key: "start", value: "2020-01-01"},
	// 	{key: "end", value: "2020-01-02"},
	// }, http.StatusOK},
	// {"post-make-reservation", "/make-reservation", "POST", []postData{
	// 	{key: "first_name", value: "Jhon"},
	// 	{key: "last_name", value: "Smith"},
	// 	{key: "email", value: "me@here.com"},
	// 	{key: "phone", value: "555-555-5555"},
	// }, http.StatusOK},
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
		}
	}
}

// test for the Reservation hander
func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "Generals's Quaters",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCtx(req)
	// add the context to the request
	req = req.WithContext(ctx)

	// RequestRecorder
	rr := httptest.NewRecorder() //NewRecorder returns an initialized ResponseRecorder.
	// so this basically simulating
	// what we get from the request response lifecycle
	// when someone fires up a web browser
	// hits our website, gets to a handler, pass that request, get a respone writer
	// and the response writer writes the response to the web browser
	// this fakes that entire process
	// or the part of it that we need for a recorder

	//now put reservation in the session
	session.Put(ctx, "reservation", reservation)

	// call the reservation handler
	// we can't call it directly
	// so turn handler reservation function into a handler
	handler := http.HandlerFunc(Repo.Reservation)

	handler.ServeHTTP(rr, req) // calling directly

	if rr.Code != http.StatusOK {
		t.Errorf("Reservation handler return wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	// ***************************************/
	// test case where reservation is not in session (reset everything)
	// reset my request
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	// get the context with the session header
	ctx = getCtx(req) // ==> have a session but doesn't have the reservation variable in it
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code : got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}
	//*********** *********/
	// test with non-existing room

	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	// get the context with the session header
	ctx = getCtx(req) // ==> have a session but doesn't have the reservation variable in it
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()
	reservation.RoomID = 100
	session.Put(ctx, "reservation", reservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code : got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

}

//{"post-make-reservation", "/make-reservation", "POST", []postData{
// 	{key: "first_name", value: "Jhon"},
// 	{key: "last_name", value: "Smith"},
// 	{key: "email", value: "me@here.com"},
// 	{key: "phone", value: "555-555-5555"},
// }, http.StatusOK},
func TestRepository_PostReservation(t *testing.T) {
	// we need to manually build a body of
	//  a post request and supply the body
	// that will pass the form data
	reqBody := "first_name=John"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=555-555-5555")

	postedData := url.Values{}
	postedData.Encode()

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx := getCtx(req) // this ctx knows about the session
	req = req.WithContext(ctx)

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, "2050-01-01") //date
	if err != nil {
		log.Println("unable to parse start-date")
	}
	endDate, err := time.Parse(layout, "2050-01-02") // date
	if err != nil {
		log.Println("unable to parse end-date!")
	}

	reservation := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    1,
		Room: models.Room{
			RoomName: "General's Quaters",
		},
	}
	session.Put(ctx, "reservation", reservation)
	rr := httptest.NewRecorder() // return *httptest.ResponseRecorder

	// set the header
	// to tell the Web Server the kind of request is comming its way
	// that is ==> form post
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler := http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)
	session.Remove(ctx, "reservation")

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code : got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// ********************************
	// test for missing POST body (failing Parse form)
	req, _ = http.NewRequest("POST", "/make-reservation", nil)
	ctx = getCtx(req) // this ctx knows about the session
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()

	session.Put(ctx, "reservation", reservation)

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for missing post body: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	//***************************************
	// test no reservation in the session

	//Session.PopString(r.Context(), "flash")
	fmt.Println("testing: no reservation in the session")
	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req) // this ctx knows about the session
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()

	//session.Put(ctx, "reservation", reservation)

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for missing post body: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// TEST :: invalid form submission

	fmt.Println("testing: invalid form submission")
	reqBody = "first_name=John"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=jhonsmith.com")
	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req) // this ctx knows about the session
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()

	session.Put(ctx, "reservation", reservation)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("PostReservation handler returned wrong response code for missing post body: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	//****************************************
	// test : failed to enter new reservation

	reqBody = "first_name=John"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=555-555-5555")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req) // this ctx knows about the session
	req = req.WithContext(ctx)

	layout = "2006-01-02"
	startDate, err = time.Parse(layout, "2050-01-01") //date
	if err != nil {
		log.Println("unable to parse start-date")
	}
	endDate, err = time.Parse(layout, "2050-01-02") // date
	if err != nil {
		log.Println("unable to parse end-date!")
	}

	reservation = models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    23,
		Room: models.Room{
			RoomName: "General's Home",
		},
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // get pain in the ass
	session.Put(ctx, "reservation", reservation)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for failed to enter new reservation: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// test : failed to enter new room-restriction

	reqBody = "first_name=John"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=555-555-5555")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req) // this ctx knows about the session
	req = req.WithContext(ctx)

	layout = "2006-01-02"
	startDate, err = time.Parse(layout, "2050-01-01") //date
	if err != nil {
		log.Println("unable to parse start-date")
	}
	endDate, err = time.Parse(layout, "2050-01-02") // date
	if err != nil {
		log.Println("unable to parse end-date!")
	}

	reservation = models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    2,
		Room: models.Room{
			RoomName: "General's Home",
		},
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // get pain in the ass
	session.Put(ctx, "reservation", reservation)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)
	fmt.Println("******Room Restriction failed********")
	fmt.Println("rr.Code:", rr.Code)
	fmt.Println("Temporary Redirect:", http.StatusTemporaryRedirect)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for failed to enter new room-restriction: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

}

func TestRepository_PostAvailability(t *testing.T) {
	// Test Pass the test
	reqBody := "start=2023-12-02"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2023-12-05")
	req, _ := http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	// this line below is important for parsing the form
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	http.HandlerFunc(Repo.PostAvailability).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("test failed for valid data : got %d expected %d", rr.Code, http.StatusOK)
	}

	// Test: fail parseform()
	req, _ = http.NewRequest("POST", "/search-availability", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	// below line fails the pasrseForm() check
	//req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	http.HandlerFunc(Repo.PostAvailability).ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("test failed for fail parseform() : got %d expected %d", rr.Code, http.StatusTemporaryRedirect)
	}
	// Test : for empty start date (Or no start date)

	reqBody = "end=2022-02-01"
	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	// this line below is important for parsing the form
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	http.HandlerFunc(Repo.PostAvailability).ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("test failed for fail parseform() : got %d expected %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// Test : for invalid end date (Or no end date)
	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader("start=2022-01-01"))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	// this line below is important for parsing the form
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	http.HandlerFunc(Repo.PostAvailability).ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("test failed for fail parseform() : got %d expected %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// Test for invalid SearchAvailabiliyForAllRooms
	reqBody = "start=2022-11-02"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2022-11-05")
	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	// this line below is important for parsing the form
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	http.HandlerFunc(Repo.PostAvailability).ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("test failed for fail parseform() : got %d expected %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// Test for len(rooms) == 0
	reqBody = "start=2023-12-02"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2023-12-02")
	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	// this line below is important for parsing the form
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	http.HandlerFunc(Repo.PostAvailability).ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("test failed for fail parseform() : got %d expected %d", rr.Code, http.StatusSeeOther)
	}

}

func TestRepository_ReservationSummary(t *testing.T) {
	// Test: for valid reservation summary
	reservation := models.Reservation{
		StartDate: time.Now(),
		EndDate:   time.Now().Add(time.Hour * 24),
		RoomID:    2,
		Room: models.Room{
			RoomName: "General's Home",
		},
	}
	req, _ := http.NewRequest("GET", "/reservation-summary", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	session.Put(ctx, "reservation", reservation)
	// this line below is important for parsing the form
	rr := httptest.NewRecorder()
	http.HandlerFunc(Repo.ReservationSummary).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("test failed valid data : got %d expected %d", rr.Code, http.StatusOK)
	}

	// Test: for invalid reservation summary
	req, _ = http.NewRequest("GET", "/reservation-summary", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	//session.Put(ctx, "reservation", reservation)
	rr = httptest.NewRecorder()
	http.HandlerFunc(Repo.ReservationSummary).ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("test failed valid data : got %d expected %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_AvailabilityJSON(t *testing.T) {
	// Test : for valid data
	req, _ := http.NewRequest("POST", "/search-availability-json", strings.NewReader("room_id=1&start=2022-01-01&end=2022-02-01"))
	ctx := getCtx(req)
	// get context with session
	req = req.WithContext(ctx)
	//set request header
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	http.HandlerFunc(Repo.AvailabilityJSON).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("valid data given but test failed: got %d, expected %d", rr.Code, http.StatusOK)
	}
	// Test: failing parseForm()
	req, _ = http.NewRequest("POST", "/search-availability-json", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	//req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	http.HandlerFunc(Repo.AvailabilityJSON).ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("test failed for not failing pasrseForm(): got %d, expected %d", rr.Code, http.StatusTemporaryRedirect)
	}
	// Test: invalid start
	req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader("room_id=1&end=2022-02-01"))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	http.HandlerFunc(Repo.AvailabilityJSON).ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("test for invalid start: got %d, expected %d", rr.Code, http.StatusTemporaryRedirect)
	}
	// Test: invalid end
	req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader("room_id=1&start=2022-02-01"))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	http.HandlerFunc(Repo.AvailabilityJSON).ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("test for invalid end: got %d, expected %d", rr.Code, http.StatusTemporaryRedirect)
	}

}

func TestRepository_BookRoom(t *testing.T) {
	// Test: for valid data
	req, _ := http.NewRequest("GET", "/book-room?id=1&sd=2022-01-01&ed=2022-02-01", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	http.HandlerFunc(Repo.BookRoom).ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("valid data given but test failed: got %d, expected %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// Test: for invalid roomID
	req, _ = http.NewRequest("GET", "/book-room?id=3&sd=2022-01-01&ed=2022-02-01", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	http.HandlerFunc(Repo.BookRoom).ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("valid data given but test failed: got %d, expected %d", rr.Code, http.StatusTemporaryRedirect)
	}

}

func TestRepository_ChooseRoom(t *testing.T) {
	// // Test: for invalid roomID
	// req, _ := http.NewRequest("GET", "/choose-room/1", nil)
	// ctx := getCtx(req)
	// req = req.WithContext(ctx)
	// req.RequestURI = "/choose-room/1"
	// rr := httptest.NewRecorder()
	// session.Put(ctx, "reservation", models.Reservation{})
	// http.HandlerFunc(Repo.ChooseRoom).ServeHTTP(rr, req)
	// if rr.Code != http.StatusSeeOther {
	// 	t.Errorf("valid data given but test failed: got %d, expected %d", rr.Code, http.StatusSeeOther)
	// }
	/*****************************************
	// first case -- reservation in session
	*****************************************/
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/choose-room/1", nil)
	ctx := getCtx(req)
	ctx = addIdToChiContext(ctx, "1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.ChooseRoom)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

var loginTests = []struct {
	name               string
	email              string
	expectedStatusCode int
	expectedHTML       string
	expectedLocation   string
}{
	{
		"valid-credentials",
		"me@here.ca",
		http.StatusSeeOther,
		"",
		"/",
	},
	// try to login with wrong credentials
	{
		"invalid-credentials",
		"jack@here.ca",
		http.StatusSeeOther,
		"",
		"/user/login",
	},
	{
		"invalid-data",
		"jack",
		http.StatusOK,
		`action="/user/login`, // that's part of the login form
		"",                    // I am not getting a location/ to go somewhere else but getting html
	},
}

func TestLogin(t *testing.T) {
	//range through all tests
	for _, e := range loginTests {
		postedData := url.Values{}
		postedData.Add("email", e.email)
		postedData.Add("password", "password")

		// create request
		req, _ := http.NewRequest("POST", "/user/login", strings.NewReader(postedData.Encode()))
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		// set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// call the handler
		handler := http.HandlerFunc(Repo.PostShowLogin)
		handler.ServeHTTP(rr, req)

		// ********* Perform Test ***********
		if rr.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		if e.expectedLocation != "" {
			// get the URL from test
			actualLoc, _ := rr.Result().Location() // this stores the location that is sent to the browser
			// check what we got == what we expected
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got location %s", e.name, e.expectedLocation, actualLoc)
			}
		}

		// checking for expected values in HTML
		if e.expectedHTML != "" {
			// read the response body into a string
			html := rr.Body.String()
			// now er have the html that is sent back by the browser
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("failed %s: expected to find %s but did not", e.name, e.expectedHTML)
			}
		}
	}
}

var adminPostShowReservationTests = []struct {
	name                 string
	url                  string
	postedData           url.Values
	expectedResponseCode int
	expectedLocation     string
	expectedHTML         string
}{
	{
		name: "valid-data-from-new",
		url:  "/admin/reservations/new/1/show",
		postedData: url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "/admin/reservations-new",
		expectedHTML:         "",
	},
	{
		name: "valid-data-from-all",
		url:  "/admin/reservations/all/1/show",
		postedData: url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "/admin/reservations-all",
		expectedHTML:         "",
	},
	{
		name: "valid-data-from-cal",
		url:  "/admin/reservations/cal/1/show",
		postedData: url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
			"year":       {"2022"},
			"month":      {"01"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "/admin/reservations-calendar?y=2022&m=01",
		expectedHTML:         "",
	},
}

// TestAdminPostShowReservation tests the AdminPostReservation handler
func TestAdminPostShowReservation(t *testing.T) {
	for _, e := range adminPostShowReservationTests {
		var req *http.Request
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/user/login", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/user/login", nil)
		}
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.RequestURI = e.url

		// set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// call the handler
		handler := http.HandlerFunc(Repo.AdminPostShowReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedResponseCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedResponseCode, rr.Code)
		}

		if e.expectedLocation != "" {
			// get the URL from test
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got location %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}

		// checking for expected values in HTML
		if e.expectedHTML != "" {
			// read the response body into a string
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("failed %s: expected to find %s but did not", e.name, e.expectedHTML)
			}
		}
	}
}

var adminPostReservationCalendarTests = []struct {
	name                 string
	postedData           url.Values
	expectedResponseCode int
	expectedLocation     string
	expectedHTML         string
	blocks               int
	reservations         int
}{
	{
		name: "cal",
		postedData: url.Values{
			"year":  {time.Now().Format("2006")},
			"month": {time.Now().Format("01")},
			fmt.Sprintf("add_block_1_%s", time.Now().AddDate(0, 0, 2).Format("2006-01-2")): {"1"},
		},
		expectedResponseCode: http.StatusSeeOther,
	},
	{
		name:                 "cal-blocks",
		postedData:           url.Values{},
		expectedResponseCode: http.StatusSeeOther,
		blocks:               1,
	},
	{
		name:                 "cal-res",
		postedData:           url.Values{},
		expectedResponseCode: http.StatusSeeOther,
		reservations:         1,
	},
}

func TestPostReservationCalendar(t *testing.T) {
	for _, e := range adminPostReservationCalendarTests {
		var req *http.Request
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/admin/reservations-calendar", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/admin/reservations-calendar", nil)
		}
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		now := time.Now()
		bm := make(map[string]int)
		rm := make(map[string]int)

		currentYear, currentMonth, _ := now.Date()
		currentLocation := now.Location()

		firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
		lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			rm[d.Format("2006-01-2")] = 0
			bm[d.Format("2006-01-2")] = 0
		}

		if e.blocks > 0 {
			bm[firstOfMonth.Format("2006-01-2")] = e.blocks
		}

		if e.reservations > 0 {
			rm[lastOfMonth.Format("2006-01-2")] = e.reservations
		}

		session.Put(ctx, "block_map_1", bm)
		session.Put(ctx, "reservation_map_1", rm)

		// set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// call the handler
		handler := http.HandlerFunc(Repo.AdminPostReservationsCalender)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedResponseCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedResponseCode, rr.Code)
		}

	}
}

var adminProcessReservationTests = []struct {
	name                 string
	queryParams          string
	expectedResponseCode int
	expectedLocation     string
}{
	{
		name:                 "process-reservation",
		queryParams:          "",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
	{
		name:                 "process-reservation-back-to-cal",
		queryParams:          "?y=2021&m=12",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
}

func TestAdminProcessReservation(t *testing.T) {
	for _, e := range adminProcessReservationTests {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/process-reservation/cal/1/do%s", e.queryParams), nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminProcessReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedResponseCode, rr.Code)
		}
	}
}

var adminDeleteReservationTests = []struct {
	name                 string
	queryParams          string
	expectedResponseCode int
	expectedLocation     string
}{
	{
		name:                 "delete-reservation",
		queryParams:          "",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
	{
		name:                 "delete-reservation-back-to-cal",
		queryParams:          "?y=2021&m=12",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
}

func TestAdminDeleteReservation(t *testing.T) {
	for _, e := range adminDeleteReservationTests {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/process-reservation/cal/1/do%s", e.queryParams), nil)
		ctx := getCtx(req)
		ctx = addIdToChiContextTest(ctx, "1", "src")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminDeleteReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedResponseCode, rr.Code)
		}
	}
}

// we need to put our reservation variable as a
// special variable into the session of the request
// using the context
func getCtx(req *http.Request) context.Context {
	// get a context
	// put our sessional variable in it
	// and store that in our request
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		fmt.Println(":( failed!")
		log.Println(err)
	}
	return ctx //==> this context actually know about the header section
	// which we need in order to read from and write to the session
}
func addIdToChiContext(parentCtx context.Context, id string) context.Context {
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", id)
	return context.WithValue(parentCtx, chi.RouteCtxKey, chiCtx)
}
func addIdToChiContextTest(parentCtx context.Context, id, src string) context.Context {
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", id)
	chiCtx.URLParams.Add("src", src)
	return context.WithValue(parentCtx, chi.RouteCtxKey, chiCtx)
}
