// here I will put any function that I want
// to be available to the DatabaseRepo interface
package dbrepo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/byt3er/bookings/internals/models"
	"golang.org/x/crypto/bcrypt"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

// 	InsertReservation inserts a reservation into the database

func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	// m.DB
	// m.App.Inproduction
	// *****************************************
	// now I have a  much safer and more robust means of
	// talking to the database.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newID int
	stmt := `insert 
				into 
					reservations (
						first_name,
						last_name,
						email,
						phone,
						start_date,
						end_date,
						room_id,
						created_at,
						updated_at)
					values( $1,
							$2,
							$3,
							$4,
							$5,
							$6,
							$7,
							$8,
							$9) returning id`
	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now()).Scan(&newID)

	if err != nil {
		fmt.Println("Reservation Not Added to the database")
		return newID, err
	}
	fmt.Println("Reservation Added to the database")
	return newID, nil
}

// InsertRoomRestriction inserts a room restriction into the database
func (m *postgresDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into room_restrictions(start_date,end_date,room_id,
				reservation_id,created_at,updated_at,restriction_id)
				values($1,$2,$3,$4,$5,$6,$7)`

	_, err := m.DB.ExecContext(ctx, stmt,
		r.StartDate,
		r.EndDate,
		r.RoomID,
		r.ReservationID,
		time.Now(),
		time.Now(),
		r.RestrictionID)
	if err != nil {
		log.Printf("Rooms Restriction not added %v", r)
		return err
	}
	log.Printf("Rooms Restriction added %v", r)
	return nil
}

// SearchAvailabilityByDatesByRoomID return true if availability exists for room id,
// and return false if no availability exists
func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	log.Println("*************")
	log.Println(start, end, roomID)
	log.Println("**************")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `
		select 
			count(id)
		from
			room_restrictions
		where
			room_id = $1
			and
			$2 < end_date and $3 > start_date;`
	// where my start_date is less than your end_date
	// and my end_date is greater than your start date.

	var numRows int

	row := m.DB.QueryRowContext(ctx, query, roomID, start, end)
	err := row.Scan(&numRows)

	if err != nil {
		return false, err
	}

	if numRows == 0 {
		// have availability
		return true, nil
	}
	// no availability
	return false, nil
}

//SearchAvailabilityForAllRooms return slice of available rooms, if any
// , for given date range
// search for availability not for a given room, but all rooms
// return whether there is availability and also returns the
// actual rooms for which there is availablity, if any for a given date range
func (m *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room
	query := `select 
				r.id, r.room_name
			from
				rooms r
			where r.id not in
			( select rr.room_id from room_restrictions rr where $1 < rr.end_date  and $2 > rr.start_date)`
	rows, err := m.DB.QueryContext(ctx, query, start, end)
	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var room models.Room
		err := rows.Scan(
			&room.ID,
			&room.RoomName,
		)
		rooms = append(rooms, room)
		if err != nil {
			return rooms, err
		}
		// check for error one more time
		if err = rows.Err(); err != nil {
			return rooms, err
		}
	}
	return rooms, err
}

// GetRoomByID gets a room by id
func (m *postgresDBRepo) GetRoomByID(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	query := `select id, room_name, created_at, updated_at from rooms where id=$1`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&room.ID,
		&room.RoomName,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	if err != nil {
		return room, err
	}
	return room, nil
}

//GetUserByID returns a user by id
func (m *postgresDBRepo) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, first_name, last_name, email, password, access_level, created_at, updated_at
				from users where id = $1`
	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.User
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return u, err
	}
	return u, nil

}

// UpdateUser update the user in the database
func (m *postgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update users set first_name = $1, last_name = $2, email =$3, access_level = $4, updated_at = $5
		where id = $6
	`
	// execute the query using our appropriate method from the database pool.
	_, err := m.DB.ExecContext(ctx, query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.AccessLevel,
		time.Now(),
		u.ID,
	)
	if err != nil {
		return err
	}
	return nil
}

// Authenticate authenticate a user
func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// create two variable to hold information from the database

	// hold the ID of the authenticated user if this return as the way they should
	var id int
	// hold the hashed password
	var hashedPassword string

	// query to find if the user exsists
	row := m.DB.QueryRowContext(ctx,
		"select id, password from users where email =$1",
		email,
	)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	// if the email is valid ,it exists in the database
	// now compare their password with mypassword

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword { // password don't match
		return 0, "", errors.New("incorrect password")
	} else if err != nil { // server error
		return 0, "", err
	}
	return id, hashedPassword, nil
}

// AllReservation returns a slice of all reservations
func (m *postgresDBRepo) AllReservation() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date,
			r.end_date, r.room_id, r.created_at, r.updated_at, 
			rm.id, rm.room_name
		from
			reservations r left join rooms rm 
		on (r.room_id = rm.id)
		order by r.start_date asc
	`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close() // other wise gone have a memory leak

	for rows.Next() {
		var i models.Reservation
		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Room.ID,
			&i.Room.RoomName,
		)
		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, i)
	}
	if err := rows.Err(); err != nil {
		return reservations, err
	}
	return reservations, nil
}

// NewReservation returns a slice of all reservations
func (m *postgresDBRepo) AllNewReservation() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date,
			r.end_date, r.room_id, r.created_at, r.updated_at, r.processed,
			rm.id, rm.room_name
		from
			reservations r left join rooms rm 
		on (r.room_id = rm.id)
		where r.processed = 0
		order by r.start_date asc
	`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close() // other wise gone have a memory leak

	for rows.Next() {
		var i models.Reservation
		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Processed,
			&i.Room.ID,
			&i.Room.RoomName,
		)
		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, i)
	}
	if err := rows.Err(); err != nil {
		return reservations, err
	}
	return reservations, nil
}

// GetReservationByID returns on reservation by ID
func (m *postgresDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date,
			r.end_date, r.room_id, r.created_at, r.updated_at, r.processed,
			rm.id, rm.room_name
		from
			reservations r left join rooms rm 
		on (r.room_id = rm.id)
		where r.ID = $1
	`
	row := m.DB.QueryRowContext(ctx, query, id)

	var r models.Reservation
	err := row.Scan(
		&r.ID,
		&r.FirstName,
		&r.LastName,
		&r.Email,
		&r.Phone,
		&r.StartDate,
		&r.EndDate,
		&r.RoomID,
		&r.CreatedAt,
		&r.UpdatedAt,
		&r.Processed,
		&r.Room.ID,
		&r.Room.RoomName,
	)

	if err != nil {
		return r, err
	}
	return r, nil
}

// UpdateReservation update a reservation in the database
func (m *postgresDBRepo) UpdateReservation(res models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update reservations set first_name = $1, last_name = $2, email =$3, phone = $4, updated_at = $5
		where id = $6
	`
	// execute the query using our appropriate method from the database pool.
	_, err := m.DB.ExecContext(ctx, query,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		time.Now(),
		res.ID,
	)
	if err != nil {
		return err
	}
	return nil
}

// DeleteReservation delete one reservation by id
func (m *postgresDBRepo) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `delete from reservations where id = $1`

	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

// UpdateProcessedForReservation updates processed for a reservation by id
func (m *postgresDBRepo) UpdateProcessedForReservation(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update reservations set processed = $1 where id = $2`

	_, err := m.DB.ExecContext(ctx, query, processed, id)
	if err != nil {
		return err
	}
	return nil
}

func (m *postgresDBRepo) AllRooms() ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `select id, room_name, created_at, updated_at from rooms order by room_name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return rooms, err
	}
	defer rows.Close()

	for rows.Next() {
		var rm models.Room
		err = rows.Scan(
			&rm.ID,
			&rm.RoomName,
			&rm.CreatedAt,
			&rm.UpdatedAt,
		)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, rm)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}
	return rooms, nil
}

// GetRestrictionsForRoomByDate returns restrictions for a room by date range
func (m *postgresDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var restrictions []models.RoomRestriction

	// inside the query "coalesce(reservation_id, 0)"
	// if reservation id has a value that value is not nul, use it
	// otherwise use zero
	query := `
		select id, coalesce(reservation_id, 0), restriction_id, room_id, start_date, end_date
		from room_restrictions where $1 < end_date and $2 >= start_date
		and room_id = $3
	`
	rows, err := m.DB.QueryContext(ctx, query, start, end, roomID)
	if err != nil {
		return restrictions, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.RoomRestriction
		err := rows.Scan(
			&r.ID,
			&r.ReservationID, // go also have null int type(note used here)
			&r.RestrictionID,
			&r.RoomID,
			&r.StartDate,
			&r.EndDate,
		)
		if err != nil {
			return restrictions, err
		}
		restrictions = append(restrictions, r)
	}
	if err = rows.Err(); err != nil {
		return restrictions, err
	}
	return restrictions, nil
}

// InsertBlockForRoom inserts a room restriction
func (m *postgresDBRepo) InsertBlockForRoom(id int, startDate time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `insert into room_restrictions (start_date, end_date, room_id, restriction_id,
				created_at, updated_at) values($1,$2,$3,$4,$5,$6)`
	// _, err := m.DB.ExecContext(ctx, query, startDate, startDate.AddDate(0, 0, 1), id, 2, time.Now(), time.Now())
	_, err := m.DB.ExecContext(ctx, query, startDate, startDate, id, 2, time.Now(), time.Now())
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// DeleteBlocksByID deletes a room restriction
func (m *postgresDBRepo) DeleteBlockByID(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `delete from room_restrictions where id = $1`
	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
