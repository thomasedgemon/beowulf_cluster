# multiprocessing_primes.py
import math
import time
from multiprocessing import Pool, cpu_count

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

def find_primes_in_range(args):
    start, end = args
    return [n for n in range(start, end) if is_prime(n)]

def chunkify(n, num_chunks):
    """Divide range 0..n into num_chunks chunks"""
    chunk_size = (n + num_chunks - 1) // num_chunks  # ceil div
    return [(i * chunk_size, min(n, (i + 1) * chunk_size)) for i in range(num_chunks)]

def main():
    N = 10_000_000
    num_workers = 16 #()  # or set manually: e.g., 4

    print(f"Using {num_workers} processes to find primes up to {N}...")
    ranges = chunkify(N, num_workers)

    start_time = time.time()
    with Pool(processes=num_workers) as pool:
        results = pool.map(find_primes_in_range, ranges)

    primes = [p for sublist in results for p in sublist]
    primes.sort()
    end_time = time.time()

    print(f"Found {len(primes)} primes up to {N}")
    print(f"Elapsed time: {end_time - start_time:.2f} seconds")

if __name__ == "__main__":
    main()
