package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/abkolan/kodex/go-projects/auction/repository"
	log "github.com/sirupsen/logrus"
)

var DB = repository.GetDB()

// Create a single random source and generator instance
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

// Generate a random number between x and x + y
func randomInRange(x, y int) int {
	return x + r.Intn(y+1) // Generate number in range [x, x+y]
}

func main() {
	// set log level to debug
	log.SetLevel(log.DebugLevel)

	db := repository.GetDB()
	defer db.Close()

	// Get the current max bid for listing id 1
	maxBid, err := getCurrentMaxBid(1)
	if err != nil {
		log.WithError(err).Error("Failed to get current max bid")
		return
	}
	log.Infof("Current max bid for listing id 1 is %d", maxBid)

	// simulate bidding n users bidding each with a random amount
	n := 10
	var wg sync.WaitGroup
	wg.Add(n)
	start := time.Now()
	for i := 0; i < n; i++ {
		go func(userId int) {
			// Simulate a user bidding with a random amount
			err := placeBid(userId)
			//wait for 500 ms
			time.Sleep(time.Duration(randomInRange(50, 500)) * time.Millisecond)

			if err != nil {
				log.WithError(err).Error("Failed to place bid")
				return
			}
			wg.Done()
		}(i)

	}
	wg.Wait()
	duration := time.Since(start)
	log.Infof("%d simulations took %s", n, duration)
}
func placeBid(userId int) error {
	//get the current max bid
	maxBid, err := getCurrentMaxBid(1)
	if err != nil {
		log.WithError(err).Error("Failed to get current max bid")
		return err
	}
	// get a random bit amount higher than maxBid
	bidAmount := randomInRange(maxBid+1, randomInRange(1, 10))
	// Set isolation level before starting the transaction
	_, err = DB.Exec("SET SESSION TRANSACTION ISOLATION LEVEL SERIALIZABLE;")
	if err != nil {
		log.WithError(err).Error("Failed to set isolation level")
		return err
	}
	_, err = DB.Exec("SET SESSION innodb_lock_wait_timeout = 5;")
	if err != nil {
		log.WithError(err).Error("Failed to set lock wait timeout")
		return err
	}
	// place the bid
	tx, err := DB.Begin()
	if err != nil {
		log.WithError(err).Error("Failed to begin transaction")
		return err
	}
	// Defer rollback in case of failure
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Step 1: Lock auction row and get current highest bid
	var maxBidAmount int
	err = tx.QueryRow("SELECT max_bid_amount FROM auction WHERE listing_id = ? FOR UPDATE", 1).Scan(&maxBidAmount)
	if err != nil {
		log.WithError(err).Error("Failed to get current max bid")
	}

	// Step 2: Check if new bid is higher than current max bid
	// Step 2: Determine bid status (accepted or rejected)
	bidStatus := "rejected"
	if bidAmount > maxBidAmount {
		bidStatus = "accepted"
	}

	// Step 3: Insert the bid into bids table (always insert, even if rejected)
	res, err := tx.Exec("INSERT INTO bids (auction_id, amount, user_id, status) VALUES (?, ?, ?, ?)", 1, bidAmount, userId, bidStatus)
	if err != nil {
		log.WithError(err).Error("Failed to insert bid")
		return err
	}

	// Step 4: If bid was accepted, update auction table with new highest bid
	if bidStatus == "accepted" {
		bidID, err := res.LastInsertId()
		if err != nil {
			log.WithError(err).Error("Failed to get last insert ID")
			return err
		}
		_, err = tx.Exec("UPDATE auction SET max_bid_id = ?, max_bid_amount = ? WHERE id = ?", bidID, bidAmount, 1)
		if err != nil {
			log.WithError(err).Error("Failed to update auction")
			return err
		}
	}
	// Step 5: Commit transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	log.Infof("Bid placed successfully! Status: %s, Amount: %d, User: %d\n", bidStatus, bidAmount, userId)
	return nil
}

func getCurrentMaxBid(auctionId int) (int, error) {
	// Query the database for the current max bid
	var maxBid int
	err := DB.QueryRow("SELECT max_bid_amount FROM auction WHERE listing_id =?", auctionId).Scan(&maxBid)

	if err != nil {
		return 0, err
	}
	return maxBid, nil
}
