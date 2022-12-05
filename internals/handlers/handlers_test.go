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
