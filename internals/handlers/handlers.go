package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/byt3er/bookings/internals/config"
	"github.com/byt3er/bookings/internals/driver"
	"github.com/byt3er/bookings/internals/forms"
	"github.com/byt3er/bookings/internals/helpers"
	"github.com/byt3er/bookings/internals/models"
	"github.com/byt3er/bookings/internals/render"
	"github.com/byt3er/bookings/internals/repository"
	"github.com/byt3er/bookings/internals/repository/dbrepo"
	"github.com/go-chi/chi"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the handler for the home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {

	//m.DB.AllUsers()
	//fmt.Println("X-Session", r.Header.Get("X-Session"))

	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	// perform some logic

	// send data to the template
	render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

// Reservation renders the make a reservation page and displays form
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("X-Session:", r.Header.Get("X-Session"))
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
		// helpers.ServerError(w, errors.New("cannot get reservation from session"))
		// return
	}

	//fmt.Println("X-Session:", r.Header.Get("X-Session"))

	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		// helpers.ServerError(w, err)
		// return
		m.App.Session.Put(r.Context(), "error", "can't find room!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.Room.RoomName = room.RoomName
	//res.Room.RoomName = "Heaven on planet Earth"

	m.App.Session.Put(r.Context(), "reservation", res)

	// convert StartDate and EndDate from time to string
	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

// PostReservation handles the submission of a  Reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "reservation not found in session!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	err := r.ParseForm()
	if err != nil {
		fmt.Println("fail to parse form!")
		m.App.Session.Put(r.Context(), "error", "can't parse form!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	form := forms.New(r.PostForm)
	//form.Has("first_name", r)
	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		reservation.FirstName = r.Form.Get("first_name")
		reservation.LastName = r.Form.Get("last_name")
		reservation.Email = r.Form.Get("email")
		fmt.Println("bitch: your form data is invalid!")
		data := make(map[string]interface{})
		data["reservation"] = reservation

		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form:      form,
			Data:      data,
			StringMap: stringMap,
		})
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

	// save the rerservation to database
	// get a new reservation_id
	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert reservation into the database!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert room_restriction into the database!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// send notifications -> first to guest
	htmlMessage := fmt.Sprintf(`
		<strong> Reservation Confirmation </strong> <br>
		Dear %s, <br>
		This is confirm your reservation from %s to %s.
	`,
		reservation.FirstName+" "+reservation.LastName,
		reservation.StartDate.Format("2006-01-02"),
		reservation.EndDate.Format("2006-01-02"),
	)

	msg := models.MailData{}
	msg.To = reservation.Email
	msg.From = "me@here.com"
	msg.Subject = "Reservation Confirmation"
	msg.Content = htmlMessage
	m.App.MailChan <- msg

	// send notifications -> first to property owner
	htmlMessage = fmt.Sprintf(`
		<strong> Reservation Notification </strong> <br>
		A reservation has been made for %s from %s to %s.
	`,
		reservation.Room.RoomName,
		reservation.StartDate.Format("2006-01-02"),
		reservation.EndDate.Format("2006-01-02"),
	)

	msg = models.MailData{}
	msg.To = "me@here.com"
	msg.From = "me@here.com"
	msg.Subject = "Reservation Notification"
	msg.Content = htmlMessage
	msg.Template = "basic.html"
	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// Generals renders the room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.tmpl", &models.TemplateData{})
}

// Majors renders the room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.tmpl", &models.TemplateData{})
}

// Availability renders the search availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

// PostAvailability handles post
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println("****** parse form failed")
		m.App.Session.Put(r.Context(), "error", "Form Error")
		http.Redirect(w, r, "/search-availability", http.StatusTemporaryRedirect)
		return
	}
	start := r.Form.Get("start") // string
	end := r.Form.Get("end")     // string

	layout := "2006-01-02"

	// parse start & end into time type
	startDate, err := time.Parse(layout, start) //date
	if err != nil {
		// helpers.ServerError(w,err)
		// return
		fmt.Println("failed to parse star-date")
		m.App.Session.Put(r.Context(), "error", "can't parse start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	endDate, err := time.Parse(layout, end) // date
	if err != nil {
		// helpers.ServerError(w, err)
		// return
		fmt.Println("failed to parese end-date")
		m.App.Session.Put(r.Context(), "error", "can't parse end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		fmt.Println("faield to query for SearchAvailablityForAllRooms")
		m.App.Session.Put(r.Context(), "error", "database error:can't find rooms!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if len(rooms) == 0 {
		// no availabliity
		fmt.Println("fail no room found.")
		m.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
		//m.App.InfoLog.Println("No Availability")
	}
	// availablity
	data := make(map[string]interface{})
	data["rooms"] = rooms

	// create a new reservation
	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// ChooseRoom display list of available rooms
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {

	// read the id {id} from the URL
	fmt.Println("chi.URLParam:", chi.URLParam(r, "id"))
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	log.Println("room id :", roomID, " : error :", err)

	if err != nil {
		m.App.Session.Put(r.Context(), "error", "invalid data!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, err)
		return
	}
	res.RoomID = roomID

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)

}

type jsonResponse struct {
	Ok        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// AvailabilityJSON handles request for availability and send JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// sd := strings.TrimSpace(r.Form.Get("start"))
	// ed := strings.TrimSpace(r.Form.Get("end"))
	sd := r.Form.Get("start")
	ed := r.Form.Get("end")
	log.Println("room_id: ", r.Form.Get("room_id"))

	log.Println("startDate: ", sd)
	log.Println("endDate: ", ed)

	layout := "2006-01-02"
	// parse start & end into time type
	startDate, err := time.Parse(layout, sd) //date
	if err != nil {
		m.App.Session.Put(r.Context(), "error", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse(layout, ed) // date
	if err != nil {
		m.App.Session.Put(r.Context(), "error", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	available, _ := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)

	resp := jsonResponse{
		Ok:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomID:    strconv.Itoa(roomID),
	}
	out, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		//log.Println(err)
		helpers.ServerError(w, err)
		return
	}
	// set the header
	// header for json file
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// Contact renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// ReservationSummary displays the reservation summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		// log.Println("cannot get item from session")
		//m.App.ErrorLog.Println("Cannot get error from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation From session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// remove the reservation out of the session
	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

// BookRoom takes URL parameter, builds a sessional variable
// and takes user to make-servation page
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	sd := r.URL.Query().Get("sd")
	ed := r.URL.Query().Get("ed")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	var res models.Reservation

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	res.RoomID = roomID
	res.StartDate = startDate
	res.EndDate = endDate
	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusTemporaryRedirect)
}

// Contact renders the contact page
func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	// Good practice
	// anytime you're doing a login logout
	// make sure to renew the token
	_ = m.App.Session.RenewToken((r.Context()))

	// parse the form
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		// Display the login page and pass the form object with errors
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	// Authenticate the user
	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "flash", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	// store the id in the session
	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

// Logout logs a user out
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	// Always good practice to renew the session token
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

// AdminAllReservations show all reservations in admin tool
func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservation()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-all-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminNewReservations shows all new reservations in admin tool
func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllNewReservation()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-new-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminShowReservation shows the reservation in the admin tool
func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	// explode and split the request URI
	// slice start form 1 not 0
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	stringMap["month"] = month
	stringMap["year"] = year

	// get reservation from the database
	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "admin-reservations-show.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

func (m *Repository) AdminPostShowReservation(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		helpers.ServerError(w, err)
		return
	}

	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]
	// stringMap := make(map[string]string)
	// stringMap["src"] = src

	year := r.Form.Get("year")
	month := r.Form.Get("month")

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = m.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	log.Println("year:", year)
	log.Println("month:", month)

	m.App.Session.Put(r.Context(), "flash", "Changes saved")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}

}

//AdminReservationsCalendar displays the reservation calendar
func (m *Repository) AdminReservationsCalender(w http.ResponseWriter, r *http.Request) {
	// assume that there is no month/ year specified
	now := time.Now()

	if r.URL.Query().Get("y") != "" { // if y is  set in the url query string
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))
		//reset now to be to given year and month in the uri
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	//months
	next := now.AddDate(0, 1, 0)  // time.Time
	last := now.AddDate(0, -1, 0) // time.Time

	nextMonth := next.Format("01")       // string
	nextMonthYear := next.Format("2006") // string

	lastMonth := last.Format("01")       // string
	lastMonthYear := last.Format("2006") // string

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	data := make(map[string]interface{})
	data["now"] = now

	// get the first and last days of the month
	currentYear, currentMonth, _ := now.Date() //returns int
	currentLocation := now.Location()
	// time.Date() works same as now.Time()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	//last day of the current month
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	// get the rooms
	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	data["rooms"] = rooms
	fmt.Printf("firstOfMonth:%s, lastOfMonth: %s", firstOfMonth, lastOfMonth)
	for _, x := range rooms {
		// create maps
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			// put default value for both the reservationMap and blockMap
			reservationMap[d.Format("2006-01-02")] = 0
			blockMap[d.Format("2006-01-02")] = 0
		}

		//************************************************
		// get all the restrictions for the current room
		//***********************************************
		restrictions, err := m.DB.GetRestrictionsForRoomByDate(x.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		//****************************
		// loop through the rerstrictions one at a time
		// determine whether its is a reservation or
		// its a block
		//********************************************
		for _, y := range restrictions {
			if y.ReservationID > 0 {

				fmt.Println("room-restrictions:", y.StartDate, ":", y.EndDate)
				// it's a reservation
				for d := y.StartDate; d.After(y.EndDate) == false; d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("2006-01-02")] = y.ReservationID
				}
			} else {
				// it's a block
				for d := y.StartDate; d.After(y.EndDate) == false; d = d.AddDate(0, 0, 1) {
					blockMap[d.Format("2006-01-02")] = y.ID
				}
				//blockMap[y.StartDate.Format("2006-01-02")] = y.ID
			}
		}

		// 1 reservationMap and blockMap for each room
		data[fmt.Sprintf("reservation_map_%d", x.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", x.ID)] = blockMap

		// the current blockMap for the current room
		// ****
		// when the calendar is first rendered to the screen, I need to store
		// the blockMap as it existed at the movement it was rendered and
		// before the user does anything to the reservation calendar
		// once they post
		// my postHandler will pull this map out of the session and
		// use that to compare what we currently have with what we used to have
		// so this way we'll be able to determine things like which blocks
		// am I getting rid of and which ones are new.
		// And I need to have this information in order to do that.
		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", x.ID), blockMap)
	}

	render.Template(w, r, "admin-reservations-calendar.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}

//AdminProcessedReservation marks a reservation as processed
func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	// set proccessed to 1 (means that reservation is processed)
	_ = m.DB.UpdateProcessedForReservation(id, 1)
	m.App.Session.Put(r.Context(), "flash", "Reservation marked as processed")
	if src == "cal" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	}

}

//AdminDeleteReservation delete a reservation by id
func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")
	// set proccessed to 1 (means that reservation is processed)
	_ = m.DB.DeleteReservation(id)
	m.App.Session.Put(r.Context(), "flash", "Reservation deleted")

	if src == "cal" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	}

}

// AdminPostReservationsCalender handles the POST of reservation calendar
func (m *Repository) AdminPostReservationsCalender(w http.ResponseWriter, r *http.Request) {
	// parse the form
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year, _ := strconv.Atoi(r.Form.Get("y"))
	month, _ := strconv.Atoi(r.Form.Get("m"))

	// process blocks
	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)
	for _, x := range rooms {
		// Get the block map from the session.
		//(block map consists of  all the blocks for a given room at the point
		// when the calendar was displayed.
		// that's the data before the user clicks submit on the calendar page)

		//***logic******
		// Loop through entire map , if we have an entry in the map
		// that does not exist in our posted data and , if the
		// restriction id > 0, then it is a block we need to remove.
		curMap := m.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", x.ID)).(map[string]int)
		for name, value := range curMap {
			// make sure the value exists in the map
			// and the value is greater than zero
			// if that's the case we have to remove it

			// ok will be false if the value is not in the map
			if val, ok := curMap[name]; ok {
				// pay attention to values > 0 , and
				// that are not in the form post(_which are unchecked)
				// the rest are just placeholders for days without blocks
				if val > 0 {
					if !form.Has(fmt.Sprintf("remove_block_%d_%s", x.ID, name)) {
						// delete the restriction by id
						//log.Println("would delete block", value)
						err := m.DB.DeleteBlockByID(value)
						if err != nil {
							log.Println(err)
						}
					}
				}

			}
		}
	}
	// now handle new blocks
	for name, _ := range r.PostForm {
		//log.Println("form has name:", name)
		// check to if prefix of a posted form element starts with "add_block_*"
		if strings.HasPrefix(name, "add_block") {
			exploded := strings.Split(name, "_")
			roomID, _ := strconv.Atoi(exploded[2])
			t, _ := time.Parse("2006-01-02", exploded[3])
			// insert a new block for roomID
			//log.Println("Would insert block for room id", roomID, "for date", exploded[3])
			err := m.DB.InsertBlockForRoom(roomID, t)
			if err != nil {
				log.Println(err)
			}
		}

	}

	m.App.Session.Put(r.Context(), "flash", "changes saved")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)

}
