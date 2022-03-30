package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
)

const DataFile = "loremipsum.txt"

// Return the word frequencies of the text argument.
//
// Split load optimally across processor cores.
func WordCount(text string) map[string]int {
	freq := make(map[string]int)
	ch := make(chan map[string]int)
	wg := new(sync.WaitGroup)
	text = strings.ToLower(text)
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	words := strings.FieldsFunc(text, f)

	n := len(words)
	size := max(1, n/10)
	for i, j := 0, size; i < n; i, j = j, j+size {
		if j > n {
			j = n
		}
		wg.Add(1)
		go func(words []string, ch chan<- map[string]int, wg *sync.WaitGroup) {
			localFreq := make(map[string]int)
			for _, word := range words {
				localFreq[word]++
			}
			ch <- localFreq
			wg.Done()
		}(words[i:j], ch, wg)
	}

	go func(ch chan map[string]int, wg *sync.WaitGroup) {
		wg.Wait()
		close(ch)
	}(ch, wg)

	for freqMap := range ch {
		for elem, key := range freqMap {
			freq[elem] += key
		}
	}

	return freq
}

func max(x, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}

// Benchmark how long it takes to count word frequencies in text numRuns times.
//
// Return the total time elapsed.
func benchmark(text string, numRuns int) int64 {
	start := time.Now()
	for i := 0; i < numRuns; i++ {
		WordCount(text)
	}
	runtimeMillis := time.Since(start).Nanoseconds() / 1e6

	return runtimeMillis
}

// Print the results of a benchmark
func printResults(runtimeMillis int64, numRuns int) {
	fmt.Printf("amount of runs: %d\n", numRuns)
	fmt.Printf("total time: %d ms\n", runtimeMillis)
	average := float64(runtimeMillis) / float64(numRuns)
	fmt.Printf("average time/run: %.2f ms\n", average)
}

func main() {
	data, err := os.ReadFile(DataFile)
	if err != nil {
		panic(err)
	}

	numRuns := 100
	runtimeMillis := benchmark(string(data), numRuns)
	printResults(runtimeMillis, numRuns)
}
