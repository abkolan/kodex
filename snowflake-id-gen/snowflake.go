package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

// Define constants for bit allocation.
const (
	// Number of bits allocated for the machine ID.
	machineIDBits uint8 = 10
	// Number of bits allocated for the sequence number.
	sequenceBits uint8 = 12

	// Calculate the maximum possible machine ID (2^10 - 1 = 1023).
	maxMachineID int64 = -1 ^ (-1 << machineIDBits)
	// Calculate the maximum possible sequence number (2^12 - 1 = 4095).
	maxSequence int64 = -1 ^ (-1 << sequenceBits)

	// Define the bit shift amount for the timestamp component.
	timestampShift = machineIDBits + sequenceBits // 10 + 12 = 22
	// Define the bit shift amount for the machine ID component.
	machineIDShift = sequenceBits // 12

	// Define a custom epoch (January 1, 2024, 00:00:00 UTC) in milliseconds.
	// You can adjust this epoch to your needs. Using a more recent epoch
	// saves bits and extends the lifespan of the 41-bit timestamp.
	// epoch int64 = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	epoch int64 = 1704067200000
)

// Snowflake struct holds the generator's state.
type Snowflake struct {
	// Mutex to protect concurrent access to shared state.
	mu sync.Mutex
	// The timestamp of the last generated ID.
	lastTimestamp int64
	// The unique ID of this machine/generator instance.
	machineID int64
	// The sequence number within the current millisecond.
	sequence int64
}

// Error returned when the clock moves backwards.
var ErrClockMovedBackwards = errors.New("clock moved backwards, refusing to generate ID")

// Error returned when an invalid machine ID is provided.
var ErrInvalidMachineID = errors.New("invalid machine ID")

// Error returned when no suitable network interface is found for machine ID generation.
var ErrNoSuitableInterface = errors.New("no suitable network interface found for machine ID")

// NewSnowflake creates and returns a new Snowflake generator instance.
// It requires a machineID (0 to maxMachineID).
func NewSnowflake(machineID int64) (*Snowflake, error) {
	// Validate the provided machineID.
	if machineID < 0 || machineID > maxMachineID {
		// Return an error if the machineID is out of the valid range.
		return nil, fmt.Errorf("%w: %d is not between 0 and %d", ErrInvalidMachineID, machineID, maxMachineID)
	}
	// Create and initialize the Snowflake struct.
	s := &Snowflake{
		// Initialize lastTimestamp to -1 to indicate no IDs generated yet.
		lastTimestamp: -1,
		// Assign the validated machineID.
		machineID: machineID,
		// Initialize sequence number to 0.
		sequence: 0,
		// The mutex is implicitly initialized (zero value is usable).
	}
	// Return the pointer to the new Snowflake instance and no error.
	return s, nil
}

// GenerateID creates and returns a new unique 64-bit Snowflake ID.
func (s *Snowflake) GenerateID() (int64, error) {
	// Lock the mutex to ensure exclusive access to shared state (thread-safety).
	s.mu.Lock()
	// Defer unlocking the mutex so it's always released, even if errors occur.
	defer s.mu.Unlock()

	// Get the current time in milliseconds since the custom epoch.
	currentTimestamp := time.Now().UnixMilli() - epoch

	// Check for clock skew (clock moving backwards).
	if currentTimestamp < s.lastTimestamp {
		// Log the event for monitoring purposes.
		log.Printf("Clock moved backwards detected. Current: %d, Last: %d", currentTimestamp, s.lastTimestamp)
		// Return an error as generating an ID could break monotonicity.
		// In a production system, you might wait or have a more complex handling strategy.
		return 0, ErrClockMovedBackwards
	}

	// If the current timestamp is the same as the last one...
	if currentTimestamp == s.lastTimestamp {
		// Increment the sequence number, applying a mask to wrap around if it exceeds maxSequence.
		s.sequence = (s.sequence + 1) & maxSequence
		// Check if the sequence number wrapped around (overflowed).
		if s.sequence == 0 {
			// Sequence overflowed, wait until the next millisecond.
			currentTimestamp = s.tilNextMillis(s.lastTimestamp)
			// Sequence is reset implicitly because it wrapped to 0 earlier.
		}
	} else {
		// If it's a new millisecond, reset the sequence number to 0.
		s.sequence = 0
	}

	// Update the last timestamp.
	s.lastTimestamp = currentTimestamp

	// Construct the 64-bit ID:
	// Shift timestamp left by the total bits of machine ID and sequence.
	// Shift machine ID left by the bits of the sequence.
	// Combine the shifted parts and the sequence using bitwise OR.
	id := (currentTimestamp << timestampShift) |
		(s.machineID << machineIDShift) |
		s.sequence

	// Return the generated ID and no error.
	return id, nil
}

// tilNextMillis blocks until the next millisecond after lastTs.
// It returns the new timestamp (milliseconds since epoch).
// This helper is called only when the sequence number overflows within a millisecond.
func (s *Snowflake) tilNextMillis(lastTs int64) int64 {
	// Get the current timestamp relative to the epoch.
	timestamp := time.Now().UnixMilli() - epoch
	// Loop as long as the current timestamp is less than or equal to the last timestamp.
	for timestamp <= lastTs {
		// Short sleep to avoid busy-waiting (optional, could also just re-read time).
		// time.Sleep(time.Microsecond * 100) // Small sleep
		// Re-fetch the current timestamp.
		timestamp = time.Now().UnixMilli() - epoch
	}
	// Return the new, distinct timestamp.
	return timestamp
}

// getMachineIDFromMAC attempts to derive a suitable machine ID from network interfaces.
func getMachineIDFromMAC() (int64, error) {
	// Get a list of all network interfaces on the system.
	interfaces, err := net.Interfaces()
	// Handle potential errors during interface retrieval.
	if err != nil {
		return 0, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	// Iterate through the list of interfaces.
	for _, iface := range interfaces {
		// Skip loopback interfaces (e.g., 'lo').
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		// Skip interfaces that are down.
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		// Get the hardware address (MAC address) of the interface.
		mac := iface.HardwareAddr
		// Skip interfaces without a valid MAC address (e.g., virtual interfaces).
		if len(mac) == 0 {
			continue
		}

		// Use the last two bytes of the MAC address for the machine ID.
		// This provides 16 bits of potential variation, which is more than
		// the 10 bits needed, reducing collision probability locally.
		// Note: This does NOT guarantee global uniqueness across datacenters
		// or different network segments without careful management.
		var derivedID int64
		// Check if MAC address has at least 2 bytes.
		if len(mac) >= 2 {
			// Combine the last two bytes into an int64.
			// Shift the second-to-last byte left by 8 bits and OR with the last byte.
			derivedID = int64(mac[len(mac)-2])<<8 | int64(mac[len(mac)-1])
		} else if len(mac) > 0 {
			// If only one byte, use that. Less ideal.
			derivedID = int64(mac[len(mac)-1])
		} else {
            // Should not happen due to earlier check, but defensive coding.
            continue
        }


		// Ensure the derived ID fits within the allocated bits using a bitmask.
		machineID := derivedID & maxMachineID

		// Log the interface name, MAC, derived ID (before mask), and final machine ID.
		log.Printf("Using interface: %s, MAC: %s, Derived ID (raw): %d, Final Machine ID: %d", iface.Name, mac.String(), derivedID, machineID)

		// Return the first valid machine ID found.
		return machineID, nil
	}

	// If no suitable interface was found after checking all, return an error.
	return 0, ErrNoSuitableInterface
}

// main is the entry point of the program.
func main() {
	// Declare machineID variable.
	var machineID int64
	// Declare error variable for handling errors during setup.
	var err error

	// Check if a command-line argument was provided.
	if len(os.Args) > 1 {
		// Attempt to parse the first argument as a 64-bit integer (base 10).
		machineID, err = strconv.ParseInt(os.Args[1], 10, 64)
		// Handle potential parsing errors.
		if err != nil {
			// Log fatal error and exit if parsing fails.
			log.Fatalf("Error parsing machine ID from command line argument '%s': %v", os.Args[1], err)
		}
		// Log the machine ID obtained from CLI.
		log.Printf("Using machine ID from command line argument: %d", machineID)
	} else {
		// If no CLI argument, attempt to derive machine ID from MAC address.
		log.Println("No command line argument for machine ID provided. Attempting to derive from MAC address...")
		machineID, err = getMachineIDFromMAC()
		// Handle potential errors during MAC address derivation.
		if err != nil {
			// Log fatal error and exit if derivation fails.
			log.Fatalf("Error deriving machine ID: %v", err)
		}
	}

	// Create a new Snowflake generator instance with the determined machineID.
	generator, err := NewSnowflake(machineID)
	// Handle potential errors during generator creation (e.g., invalid machine ID).
	if err != nil {
		// Log fatal error and exit if generator creation fails.
		log.Fatalf("Error creating Snowflake generator: %v", err)
	}

	// Log successful generator creation.
	log.Printf("Snowflake generator created successfully with Machine ID: %d", machineID)
	log.Println("Generating 10 Snowflake IDs...")

	// Generate and print 10 example IDs.
	for i := 0; i < 10; i++ {
		// Call GenerateID to get a new ID.
		id, err := generator.GenerateID()
		// Handle potential errors during ID generation (e.g., clock skew).
		if err != nil {
			// Log error if ID generation fails.
			log.Printf("Error generating ID %d: %v", i+1, err)
			// Continue to the next iteration in case of transient errors.
			continue
		}
		// Print the generated ID.
		fmt.Println(id)

		// Optional: Add a small delay to see sequence numbers reset or increment.
		// time.Sleep(time.Millisecond * 10)
	}

	// Example: Generate IDs concurrently using goroutines
	n := 20000
	log.Println("\nGenerating 5 IDs concurrently...")
	// Use a WaitGroup to wait for all goroutines to finish.
	var wg sync.WaitGroup
	// Launch n goroutines.
	// Print current time with nanoseconds in human readable format.
	log.Printf("Current time: %v", time.Now().Format(time.RFC3339Nano))
	// Record the start time to measure concurrent generation duration.
	start := time.Now()
	for i := 0; i < n; i++ {
		// Increment the WaitGroup counter for each goroutine.
		wg.Add(1)
		// Launch a goroutine.
		go func(workerNum int) {
			// Defer Done() to decrement the counter when the goroutine exits.
			defer wg.Done()
			// Generate an ID within the goroutine.
			id, err := generator.GenerateID()
			// Handle potential errors.
			if err != nil {
				fmt.Printf("[Worker %d] Error generating ID: %v", workerNum, err)
				return
			}
			// Print the ID along with the worker number.
			fmt.Printf("[Worker %d] Generated ID: %d", workerNum, id)
		}(i) // Pass loop variable i to avoid closure issues
	}	
	// Wait for all goroutines launched by this loop to finish.
	wg.Wait()
	// Calculate and log the duration of concurrent generation.
	log.Printf("Concurrent generation took: %v", time.Since(start))
	// Print current time with nanoseconds in human readable format.
	log.Printf("Current time: %v", time.Now().Format(time.RFC3339Nano))
	// Indicate completion of concurrent generation.
	log.Println("Concurrent generation finished.")
}