package repository

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/bxcodec/faker/v4"
	log "github.com/sirupsen/logrus"
)

// User represents a user in the system
type User struct {
	ID   int
	Name string
}

// UserRepository handles user data interactions
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new repository
func NewUserRepository() *UserRepository {
	return &UserRepository{
		db: GetDB(),
	}
}

func (u *UserRepository) FillFakes(n int) {
	// Insert n fake users
	for i := 0; i < n; i++ {
		name := faker.Name() // Generates a random name
		_, err := u.db.Exec("INSERT INTO users (name) VALUES (?)", name)
		if err != nil {
			log.Error("Failed to insert user:", err)
		} else {
			log.Info("Inserted user:", name)
		}
	}
}

func (u *UserRepository) GetAllUsers() ([]User, error) {
	rows, err := u.db.Query("SELECT id, name FROM users ORDER by id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

type Seat struct {
	ID     int
	Name   string
	UserID int
	TripID int
}

type SeatRepository struct {
	db *sql.DB
}

// NewSeatRepository creates a new repository
func NewSeatRepository() *SeatRepository {
	return &SeatRepository{
		db: GetDB(),
	}
}

func (s *SeatRepository) CreateEmptySeats() {
	for i := 1; i <= 20; i++ {
		name := fmt.Sprintf("%d-A", i)
		_, err := s.db.Exec("INSERT INTO seats (name) VALUES (?)", name)
		if err != nil {
			log.Error("Failed to insert seat:", err)
		} else {
			log.Info("Inserted seat:", name)
		}

		name = fmt.Sprintf("%d-B", i)
		_, err = s.db.Exec("INSERT INTO seats (name) VALUES (?)", name)
		if err != nil {
			log.Error("Failed to insert seat:", err)
		} else {
			log.Info("Inserted seat:", name)
		}

		name = fmt.Sprintf("%d-C", i)
		_, err = s.db.Exec("INSERT INTO seats (name) VALUES (?)", name)
		if err != nil {
			log.Error("Failed to insert seat:", err)
		} else {
			log.Info("Inserted seat:", name)
		}

		name = fmt.Sprintf("%d-D", i)
		_, err = s.db.Exec("INSERT INTO seats (name) VALUES (?)", name)
		if err != nil {
			log.Error("Failed to insert seat:", err)
		} else {
			log.Info("Inserted seat:", name)
		}

		name = fmt.Sprintf("%d-E", i)
		_, err = s.db.Exec("INSERT INTO seats (name) VALUES (?)", name)
		if err != nil {
			log.Error("Failed to insert seat:", err)
		} else {
			log.Info("Inserted seat:", name)
		}

		name = fmt.Sprintf("%d-F", i)
		_, err = s.db.Exec("INSERT INTO seats (name) VALUES (?)", name)
		if err != nil {
			log.Error("Failed to insert seat:", err)
		} else {
			log.Info("Inserted seat:", name)
		}
	}
}

func (s *SeatRepository) GetAllSeats() ([]Seat, error) {
	rows, err := s.db.Query("SELECT id, name, COALESCE(user_id,-1), COALESCE(trip_id,-1) FROM seats order by id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seats []Seat
	for rows.Next() {
		var seat Seat
		err := rows.Scan(&seat.ID, &seat.Name, &seat.UserID, &seat.TripID)
		if err != nil {
			return nil, err
		}
		seats = append(seats, seat)
	}
	return seats, nil
}

func (s *SeatRepository) PrintSeatMap() {
	seats, err := s.GetAllSeats()
	if err != nil {
		log.Error("Failed to get seats:", err)
		return
	}

	// create a 3x20 bool array
	var seatMap [7][21]bool

	for _, seat := range seats {
		//read seat name and convert it to the bool array as above
		seatName := seat.Name
		seats := strings.Split(seatName, "-")
		seatCol, _ := strconv.Atoi(seats[0])
		seatRow := letterToNumber(seats[1])
		//mark the seat as occupied if user_id is not null
		if seat.UserID != -1 {
			seatMap[seatRow][seatCol] = true
		}
	}
	//print the seat map
	for i := 1; i <= 6; i++ {
		if i == 4 {
			fmt.Println()
		}
		for j := 1; j <= 20; j++ {
			if seatMap[i][j] {
				fmt.Print(" x ")
			} else {
				fmt.Print(" . ")
			}
		}
		fmt.Println()
	}

}

func (s *SeatRepository) ResetAllSeats() {
	_, err := s.db.Exec("UPDATE seats SET user_id = NULL, trip_id = 1")
	if err != nil {
		log.Error("Failed to reset all seats:", err)
		return
	}
}

func letterToNumber(letter string) int {
	letter = strings.ToUpper(letter) // Convert to uppercase to handle lowercase letters
	if len(letter) != 1 || letter[0] < 'A' || letter[0] > 'Z' {
		return -1 // Return -1 for invalid inputs
	}
	return int(letter[0]-'A') + 1
}

//func (s *SeatRepository) GetAllSeats()

type AirlineRepository struct {
	db *sql.DB
}

func NewAirlineRepository() *AirlineRepository {
	return &AirlineRepository{
		db: GetDB(),
	}
}

func (a *AirlineRepository) CreateAirline() {

}

func (a *AirlineRepository) Initialize() {
	// Fill Seats
	seatsRepo := NewSeatRepository()
	seatsRepo.CreateEmptySeats()

	// Fill Users
	userRepo := NewUserRepository()
	userRepo.FillFakes(120)

	// Create a Trip
	//TODO: Create a trip
}
