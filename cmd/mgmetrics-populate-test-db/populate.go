// mgmetrics-populate-text-db is used for populating the test database
// It does this by simulating a number of streams of incoming metrics
// Usage: mgmetrics-populate-text-db -s 10 -length 100
// Where 10 is the number of streams, and 100 would be the length of each stream
// The default is 10 streams with 10000 metrics per stream (length)

package main

import (
	"flag"
	"fmt"
	"github.com/joegoldbeck/mgmetrics/models"
	"math/rand"
	"sync"
	"time"
)

func main() {
	// Parse input arguments
	var numStreams int
	var streamLength int64
	flag.IntVar(&numStreams, "streams", 10, "determines the number of streams")
	flag.Int64Var(&streamLength, "length", 10000, "determines the length of stream in ms")
	flag.Parse()

	// Reset the database. This line here should never go near a production application
	models.DropDB("host=localhost port=5432 user=metrics_user dbname=postgres password=dev_only sslmode=disable")

	// Connect to the database
	db, err := models.OpenDB("host=localhost port=5432 user=metrics_user dbname=postgres password=dev_only sslmode=disable")

	if err != nil {
		fmt.Println("Error connecting to database")
		return
	}
	defer db.Close()

	// Start the streams
	var wg sync.WaitGroup
	wg.Add(numStreams)

	for i := 0; i < numStreams; i++ {
		metricIterator := RandomMetricsIterator(streamLength)

		go func() {
			defer wg.Done()
			for m := range metricIterator {
				if _, err := db.AddMetric(m); err != nil {
					// wait a moment if there is an error
					time.Sleep(100 * time.Millisecond)
					db.AddMetric(m)
				}
				// wait a tiny bit regardless not overwhelm postgres too much
				time.Sleep(100 * time.Microsecond)
			}
		}()
		// wait a brief moment between starting streams
		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()
	fmt.Printf("Done!")
}

// RandomMetricsIterator is a channel-based iterator returning PlainMetrics
// Within an iterator, the metrics will have the same key and tags,
// and the value will change via a random walk
func RandomMetricsIterator(length int64) <-chan models.PlainMetric {
	ch := make(chan models.PlainMetric)
	timeIterator := TimeIterator(length)
	randomWalk := RandomWalkFloatIterator()
	key := RandomKey()
	tags := RandomTags()

	go func() {
		for t := range timeIterator {
			ch <- models.PlainMetric{Key: key, Tags: tags, Timestamp: t, Value: <-randomWalk}
		}
		close(ch)
	}()
	return ch
}

// Int64Iterator is a channel-based iterator returning int64 within a certain range
func Int64Iterator(min int64, max int64) <-chan int64 {
	ch := make(chan int64)
	go func() {
		for i := min; i < max; i++ {
			ch <- i
		}
		close(ch)
	}()
	return ch
}

// TimeIterator is a channel-based iterator returning a certain number of reasonable int64 timestamps
func TimeIterator(length int64) <-chan int64 {
	start := 1505327537848 + rand.Int63n(10000000000)
	return Int64Iterator(start, start+length)
}

// RandomWalkFloatIterator is a channel-based iterator returning float64 numbers which start at a random point and then random walk from there
func RandomWalkFloatIterator() <-chan float64 {
	ch := make(chan float64)
	val := rand.Float64() * 100
	go func() {
		for {
			ch <- val + rand.Float64() - 0.5
		}
	}()
	return ch
}

// RandomTags returns a random tags room-x, bed-y, patient-z, where x,y,z are random small ints
func RandomTags() []string {
	return []string{
		fmt.Sprintf("%s-%d", "room", rand.Intn(100)),
		fmt.Sprintf("%s-%d", "bed", rand.Intn(100)),
		fmt.Sprintf("%s-%d", "patient", rand.Intn(100))}
}

// RandomKey returns a random key that might be a medical measurement
func RandomKey() string {
	var possibleKeys = []string{"heartrate", "blood pressure", "temperature", "respiratory rate", "pulse oximetry"}
	return possibleKeys[rand.Intn(len(possibleKeys))]
}
