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
		r.ReservatinID,
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
		"select id, password form users where email =$1",
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
