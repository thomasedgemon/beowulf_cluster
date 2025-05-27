package main

//2.29 seconds
import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"
)

const maxNumber = 10_000_000

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

func findPrimesParallel(limit int, numWorkers int) {
	jobs := make(chan int, 1000)
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for num := range jobs {
				_ = isPrime(num)
			}
		}()
	}

	for i := 2; i < limit; i++ {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
}

func main() {
	numCores := runtime.NumCPU()
	runtime.GOMAXPROCS(numCores)

	start := time.Now()
	findPrimesParallel(maxNumber, numCores)
	elapsed := time.Since(start)

	fmt.Printf("Completed on %d cores. Total time: %v\n", numCores, elapsed)
}
