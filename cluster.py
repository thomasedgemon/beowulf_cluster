# mpi_primes.py
from mpi4py import MPI
import math
import time

def is_prime(n):
    if n < 2:
        return False
    if n == 2:
        return True
    if n % 2 == 0:
        return False
    for i in range(3, int(math.isqrt(n)) + 1, 2):
        if n % i == 0:
            return False
    return True

def find_primes_in_range(start, end):
    return [n for n in range(start, end) if is_prime(n)]

def main():
    comm = MPI.COMM_WORLD
    rank = comm.Get_rank()
    size = comm.Get_size()

    N = 1_000_000

    if rank == 0:
        print(f"Using {size} MPI processes to find primes up to {N}...")
        start_time = time.time()

    # Divide the work
    numbers_per_rank = (N + size - 1) // size  # ceil division
    start = rank * numbers_per_rank
    end = min((rank + 1) * numbers_per_rank, N)

    local_primes = find_primes_in_range(start, end)

    all_primes = comm.gather(local_primes, root=0)

    if rank == 0:
        primes = [p for sublist in all_primes for p in sublist]
        primes.sort()
        end_time = time.time()

        print(f"Found {len(primes)} primes up to {N}")
        print(f"Elapsed time: {end_time - start_time:.2f} seconds")

if __name__ == "__main__":
    main()