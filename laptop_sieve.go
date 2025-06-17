package main

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"
)

const maxNumber = 500_000_000

// Bit manipulation helpers
func setBit(arr []uint64, i int) {
	arr[i/64] |= 1 << (i % 64)
}

func clearBit(arr []uint64, i int) {
	arr[i/64] &^= 1 << (i % 64)
}

func getBit(arr []uint64, i int) bool {
	return (arr[i/64]>>(i%64))&1 == 1
}

// Generate small primes up to sqrt(maxNumber) (skipping even numbers)
func smallPrimes(limit int) []int {
	isPrime := make([]bool, limit+1)
	for i := 2; i <= limit; i++ {
		isPrime[i] = true
	}
	for p := 3; p*p <= limit; p += 2 {
		if isPrime[p] {
			for multiple := p * p; multiple <= limit; multiple += 2 * p {
				isPrime[multiple] = false
			}
		}
	}
	primes := []int{2}
	for i := 3; i <= limit; i += 2 {
		if isPrime[i] {
			primes = append(primes, i)
		}
	}
	return primes
}

// Segmented sieve using bit-packed array, only for odd numbers
func sieveSegment(start, end int, basePrimes []int) ([]uint64, int) {
	if start%2 == 0 {
		start++
	}
	count := (end - start + 1) / 2
	bits := make([]uint64, (count+63)/64)
	for i := range bits {
		bits[i] = ^uint64(0) // all bits set = all primes assumed
	}
	for _, p := range basePrimes {
		if p == 2 {
			continue
		}
		multiple := p * p
		if multiple < start {
			multiple = ((start + p - 1) / p) * p
		}
		if multiple%2 == 0 {
			multiple += p
		}
		for j := multiple; j < end; j += 2 * p {
			index := (j - start) / 2
			clearBit(bits, index)
		}
	}
	// Count primes
	total := 0
	for i := 0; i < count; i++ {
		if getBit(bits, i) {
			total++
		}
	}
	return bits, total
}

// Parallel segment sieving
func findPrimesParallel(start, end, numWorkers int, basePrimes []int) int {
	if start%2 == 0 {
		start++
	}
	rangeSize := (end - start) / numWorkers
	var wg sync.WaitGroup
	totalCount := 0
	var mu sync.Mutex
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			s := start + w*rangeSize
			e := s + rangeSize
			if w == numWorkers-1 {
				e = end
			}
			if s%2 == 0 {
				s++
			}
			_, count := sieveSegment(s, e, basePrimes)
			mu.Lock()
			totalCount += count
			mu.Unlock()
		}(w)
	}
	wg.Wait()
	return totalCount
}

func main() {
	numCores := runtime.NumCPU()
	runtime.GOMAXPROCS(numCores)

	basePrimes := smallPrimes(int(math.Sqrt(float64(maxNumber))))

	start := time.Now()
	totalCount := findPrimesParallel(1, maxNumber, numCores, basePrimes)
	elapsed := time.Since(start)

	fmt.Printf("Computation completed. Time: %v\n", elapsed)
	fmt.Printf("Total primes found under %d: %d\n", maxNumber, totalCount+1) // +1 for prime 2
}