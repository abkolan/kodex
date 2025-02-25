package main

import (
	"sync"
	"time"

	"github.com/abkolan/kodex/go-projects/airline/repository"
	log "github.com/sirupsen/logrus"
)

func book(user *repository.User) (*repository.Seat, error) {
	DB := repository.GetDB()
	txn, _ := DB.Begin()

	row := txn.QueryRow(`SELECT id,name,trip_id FROM seats
						WHERE trip_id = 1 AND user_id IS NULL 
						ORDER BY id LIMIT 1 FOR UPDATE SKIP LOCKED`)
	if row.Err() != nil {
		log.Error("Failed to get seat:", row.Err())
		return nil, row.Err()
	}
	var seat repository.Seat
	err := row.Scan(&seat.ID, &seat.Name, &seat.TripID)
	if err != nil {
		log.Error("Failed to scan seat:", err)
		return nil, err
	}
	_, err = txn.Exec(`UPDATE seats SET user_id = ? WHERE id = ?`, user.ID, seat.ID)
	if err != nil {
		log.Error("Failed to update seat:", err)
		return nil, err
	}
	// commit transaction
	err = txn.Commit()
	if err != nil {
		log.Error("Failed to commit transaction:", err)
		return nil, err
	}
	log.Infof("User %s booked seat %s", user.Name, seat.Name)

	return &seat, nil
}

func main() {
	// set log level to debug
	log.SetLevel(log.DebugLevel)

	db := repository.GetDB()
	defer db.Close()

	// reset all seats
	seatRepo := repository.NewSeatRepository()
	seatRepo.ResetAllSeats()

	// get all users
	userRepo := repository.NewUserRepository()
	users, err := userRepo.GetAllUsers()
	if err != nil {
		log.Error("Failed to get users:", err)
		return
	}
	log.Debugf("simulating %d users", len(users))

	var wg sync.WaitGroup
	wg.Add(len(users))
	start := time.Now()
	for ix := range users {
		go func(user *repository.User) {
			//book a seat for the user
			seat, err := book(user)
			if err != nil {
				log.Error("Failed to book seat:", err)
			} else {
				log.Infof("User %s booked seat %s", user.Name, seat.Name)
			}
			wg.Done()
		}(&users[ix])
	}
	wg.Wait()
	duration := time.Since(start)
	log.Infof("Booking took %s", duration)
	log.Info("Seat map after booking")
	seatRepo.PrintSeatMap()
}
