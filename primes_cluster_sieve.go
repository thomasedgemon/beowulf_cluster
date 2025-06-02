package main

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	mpi "github.com/mnlphlp/gompi"
)

const maxNumber = 10_000_000

// Simple sieve to get small primes up to sqrt(n)
func smallPrimes(limit int) []int {
	isPrime := make([]bool, limit+1)
	for i := 2; i <= limit; i++ {
		isPrime[i] = true
	}
	for p := 2; p*p <= limit; p++ {
		if isPrime[p] {
			for multiple := p * p; multiple <= limit; multiple += p {
				isPrime[multiple] = false
			}
		}
	}
	primes := []int{}
	for i, v := range isPrime {
		if v {
			primes = append(primes, i)
		}
	}
	return primes
}

// Segmented sieve for a subrange [start, end)
func sieveSegment(start, end int, basePrimes []int) []bool {
	size := end - start
	isPrime := make([]bool, size)
	for i := range isPrime {
		isPrime[i] = true
	}
	for _, p := range basePrimes {
		// Find the smallest multiple of p in [start, end)
		multiple := ((start + p - 1) / p) * p
		if multiple < p*p {
			multiple = p * p
		}
		for j := multiple; j < end; j += p {
			isPrime[j-start] = false
		}
	}
	return isPrime
}

// Parallel segment sieving
func findPrimesParallel(start, end, numWorkers int, basePrimes []int) {
	rangeSize := (end - start) / numWorkers
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			s := start + w*rangeSize
			e := s + rangeSize
			if w == numWorkers-1 {
				e = end
			}
			_ = sieveSegment(s, e, basePrimes)
		}(w)
	}
	wg.Wait()
}

func main() {
	mpi.Init()
	defer mpi.Finalize()

	comm := mpi.NewComm(false)
	rank := comm.GetRank()
	size := comm.GetSize()

	numCores := runtime.NumCPU()
	runtime.GOMAXPROCS(numCores)

	rangeSize := maxNumber / size
	startRange := rank * rangeSize
	endRange := (rank + 1) * rangeSize
	if rank == size-1 {
		endRange = maxNumber
	}

	// All processes compute the base primes up to sqrt(maxNumber)
	basePrimes := smallPrimes(int(math.Sqrt(float64(maxNumber))))

	if rank == 0 {
		start := time.Now()

		// Rank 0 also does its own work
		findPrimesParallel(startRange, endRange, numCores, basePrimes)

		// Wait for other ranks to finish
		for i := 1; i < size; i++ {
			var buf []int
			comm.Recv(&buf, i, 0)
		}

		elapsed := time.Since(start)
		fmt.Printf("All workers completed. Total time: %v\n", elapsed)

	} else {
		findPrimesParallel(startRange, endRange, numCores, basePrimes)
		comm.Send([]int{1}, 0, 0)
	}
}
