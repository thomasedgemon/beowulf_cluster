package main

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
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

func findPrimesInRange(start, end int, wg *sync.WaitGroup, ch chan []int) {
	defer wg.Done()
	primes := []int{}
	for n := start; n < end; n++ {
		if isPrime(n) {
			primes = append(primes, n)
		}
	}
	ch <- primes
}

func chunkify(n, numChunks int) [][2]int {
	chunkSize := (n + numChunks - 1) / numChunks // ceil div
	chunks := make([][2]int, 0, numChunks)
	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := (i + 1) * chunkSize
		if end > n {
			end = n
		}
		chunks = append(chunks, [2]int{start, end})
	}
	return chunks
}

func main() {
	N := 10000000
	numWorkers := 16 // or runtime.NumCPU()

	fmt.Printf("Using %d goroutines to find primes up to %d...\n", numWorkers, N)

	ranges := chunkify(N, numWorkers)
	startTime := time.Now()

	var wg sync.WaitGroup
	ch := make(chan []int, numWorkers)

	for _, r := range ranges {
		wg.Add(1)
		go findPrimesInRange(r[0], r[1], &wg, ch)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var allPrimes []int
	for primes := range ch {
		allPrimes = append(allPrimes, primes...)
	}

	sort.Ints(allPrimes)

	endTime := time.Now()
	fmt.Printf("Found %d primes up to %d\n", len(allPrimes), N)
	fmt.Printf("Elapsed time: %.2f seconds\n", endTime.Sub(startTime).Seconds())
}
