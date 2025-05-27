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

func findPrimesParallel(start, end, numWorkers int) {
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

	go func() {
		for i := start; i < end; i++ {
			jobs <- i
		}
		close(jobs)
	}()

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

	if rank == 0 {
		start := time.Now() // ⏱️ Start timer

		for i := 1; i < size; i++ {
			var buf []int
			comm.Recv(&buf, i, 0)
		}

		elapsed := time.Since(start) // ⏱️ Measure elapsed
		fmt.Printf("All workers completed. Total time: %v\n", elapsed)

	} else {
		findPrimesParallel(startRange, endRange, numCores)
		comm.Send([]int{1}, 0, 0)
	}
}