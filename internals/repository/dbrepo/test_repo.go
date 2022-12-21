package dbrepo

import (
	"errors"
	"fmt"
	"time"

	"github.com/byt3er/bookings/internals/models"
)

func (m *testDBRepo) AllUsers() bool {
	return true
}

// 	InsertReservation inserts a reservation into the database

func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	// m.DB
	if res.RoomID == 1 {
		return 1, nil
	} else if res.RoomID == 2 {
		return 2, nil
	} else if res.RoomID > 2 {
		return 0, errors.New("error: failed to enter new reservation")
	}
	return 0, nil
}

// InsertRoomRestriction inserts a room restriction into the database
func (m *testDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	fmt.Println("roomID:", r.RoomID)
	if r.ReservationID == 2 {
		return errors.New("some error")
	}
	return nil
}

// SearchAvailabilityByDatesByRoomID return true if availability exists for room id,
// and return false if no availability exists
func (m *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {

	return false, nil
}

//SearchAvailabilityForAllRooms return slice of available rooms, if any
// , for given date range
// search for availability not for a given room, but all rooms
// return whether there is availability and also returns the
// actual rooms for which there is availablity, if any for a given date range
func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {

	var rooms []models.Room
	today := time.Now()
	if start.Before(today) {
		return rooms, errors.New("DB error")
	}

	if start == end {
		return rooms, nil
	}

	rooms = append(rooms, models.Room{
		ID:       1,
		RoomName: "some-room",
	})

	return rooms, nil
}

// GetRoomByID gets a room by id
func (m *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	var room models.Room
	if id > 2 {
		return room, errors.New("some error")
	}
	return room, nil
}

//GetUserByID returns a user by id
func (m *testDBRepo) GetUserByID(id int) (models.User, error) {
	var u models.User
	return u, nil

}

// UpdateUser update the user in the database
func (m *testDBRepo) UpdateUser(u models.User) error {

	return nil
}

// Authenticate authenticate a user
func (m *testDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	if email == "me@here.ca" {
		return 1, "", nil
	}
	return 0, "", errors.New("some error")
}
func (m *testDBRepo) AllReservation() ([]models.Reservation, error) {
	var reservations []models.Reservation
	return reservations, nil
}
func (m *testDBRepo) AllNewReservation() ([]models.Reservation, error) {
	var reservations []models.Reservation
	return reservations, nil
}
func (m *testDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	var r models.Reservation
	return r, nil
}

func (m *testDBRepo) UpdateReservation(r models.Reservation) error {
	return nil
}
func (m *testDBRepo) DeleteReservation(id int) error {

	return nil

}
func (m *testDBRepo) UpdateProcessedForReservation(id, processed int) error {

	return nil
}

func (m *testDBRepo) AllRooms() ([]models.Room, error) {
	var rooms []models.Room

	return rooms, nil
}
func (m *testDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	var restrictions []models.RoomRestriction

	return restrictions, nil
}

// InsertBlockForRoom inserts a room restriction
func (m *testDBRepo) InsertBlockForRoom(id int, startDate time.Time) error {

	return nil
}

// DeleteBlocksByID deletes a room restriction
func (m *testDBRepo) DeleteBlockByID(id int) error {

	return nil
}
