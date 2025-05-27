package main

//for the nodes
import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"time"

	mpi "github.com/mnlphlp/gompi"
)

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}
	sqrtN := int(math.Sqrt(float64(n)))
	for i := 3; i <= sqrtN; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func findPrimes(start, end int) []int {
	if start < 2 {
		start = 2
	}

	// Use all 4 cores within each worker
	numCores := 4
	chunkSize := (end - start + 1) / numCores

	results := make(chan []int, numCores)

	// Launch goroutines for each core
	for i := 0; i < numCores; i++ {
		coreStart := start + i*chunkSize
		coreEnd := coreStart + chunkSize - 1
		if i == numCores-1 {
			coreEnd = end // Last chunk gets remainder
		}

		go func(s, e int) {
			var primes []int
			for n := s; n <= e; n++ {
				if isPrime(n) {
					primes = append(primes, n)
				}
			}
			results <- primes
		}(coreStart, coreEnd)
	}

	// Collect results from all cores
	var allPrimes []int
	for i := 0; i < numCores; i++ {
		coreResult := <-results
		allPrimes = append(allPrimes, coreResult...)
	}

	sort.Ints(allPrimes) // Sort since cores may finish out of order
	return allPrimes
}

func encodeInts(data []int) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		panic(fmt.Sprintf("Encoding error: %v", err))
	}
	return buf.Bytes()
}

func decodeInts(data []byte) []int {
	var result []int
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&result)
	if err != nil {
		panic(fmt.Sprintf("Decoding error: %v", err))
	}
	return result
}

func main() {
	mpi.Init()
	defer mpi.Finalize()

	comm := mpi.NewComm(false) // false means don't panic on error
	rank := comm.GetRank()
	size := comm.GetSize()

	// Ensure we have exactly 5 processes (1 master + 4 workers)
	if size != 5 {
		if rank == 0 {
			fmt.Printf("Error: This program requires exactly 5 MPI processes (1 master + 4 workers)\n")
			fmt.Printf("Current size: %d\n", size)
			fmt.Printf("Run with: mpirun -np 5 ./mpi_primes\n")
		}
		return
	}

	// Get max number from CLI
	N := 10000000
	if len(os.Args) > 1 {
		if val, err := strconv.Atoi(os.Args[1]); err == nil {
			N = val
		}
	}

	if rank == 0 {
		// MASTER NODE - Does no computational work
		overallStart := time.Now()
		fmt.Printf("Master: Starting prime search up to %d using 4 worker nodes\n", N)
		fmt.Printf("Master: Waiting for results from workers...\n")

		var allPrimes []int

		// Collect results from all 4 workers
		for workerRank := 1; workerRank <= 4; workerRank++ {
			fmt.Printf("Master: Waiting for results from worker %d\n", workerRank)

			var receivedData []int
			comm.Recv(&receivedData, workerRank, 0)

			fmt.Printf("Master: Received %d primes from worker %d\n", len(receivedData), workerRank)
			allPrimes = append(allPrimes, receivedData...)
		}

		// Sort all collected primes
		fmt.Printf("Master: Sorting %d total primes...\n", len(allPrimes))
		sort.Ints(allPrimes)

		elapsed := time.Since(overallStart)

		// Display results
		fmt.Printf("\n=== RESULTS ===\n")
		fmt.Printf("Found %d primes up to %d\n", len(allPrimes), N)
		fmt.Printf("Total elapsed time: %.2f seconds\n", elapsed.Seconds())

		if len(allPrimes) > 0 {
			fmt.Printf("First 10 primes: ")
			end := 10
			if len(allPrimes) < 10 {
				end = len(allPrimes)
			}
			for i := 0; i < end; i++ {
				fmt.Printf("%d ", allPrimes[i])
			}
			fmt.Println()

			if len(allPrimes) > 10 {
				fmt.Printf("Last 10 primes: ")
				start := len(allPrimes) - 10
				if start < 0 {
					start = 0
				}
				for i := start; i < len(allPrimes); i++ {
					fmt.Printf("%d ", allPrimes[i])
				}
				fmt.Println()
			}
		}

	} else {
		// WORKER NODES (ranks 1, 2, 3, 4)
		workerID := rank
		numWorkers := 4

		// Divide the work among 4 workers
		chunkSize := (N + numWorkers - 1) / numWorkers
		start := (workerID-1)*chunkSize + 1 // workerID-1 because workers are ranks 1-4
		end := workerID * chunkSize

		// Adjust the first worker to start from 2 (first prime)
		if workerID == 1 && start < 2 {
			start = 2
		}

		// Make sure we don't exceed N
		if end > N {
			end = N
		}

		fmt.Printf("Worker %d: Searching for primes in range [%d, %d] using 4 cores\n", workerID, start, end)

		startTime := time.Now()
		localPrimes := findPrimes(start, end)
		duration := time.Since(startTime)

		fmt.Printf("Worker %d: Found %d primes in %.2f seconds\n",
			workerID, len(localPrimes), duration.Seconds())

		// Send results back to master
		fmt.Printf("Worker %d: Sending %d primes to master\n", workerID, len(localPrimes))
		comm.Send(localPrimes, 0, 0)

		fmt.Printf("Worker %d: Complete\n", workerID)
	}
}
