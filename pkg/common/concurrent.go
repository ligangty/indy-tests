package common

import (
	"fmt"
	"sync"
)

func ConcurrentRun(numWorkers int, artifacts map[string][]string, job func(md5, originalURL, targetURL string) bool) bool {
	fmt.Printf("Start to run job in concurrent mode with thread number %v\n", numWorkers)
	ch := make(chan []string, numWorkers*5) // This buffered number of chan can be anything as long as it's larger than numWorkers
	var wg sync.WaitGroup
	var mu sync.Mutex
	var results = []bool{}

	// This starts numWorkers number of goroutines that wait for something to do
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			for {
				a, ok := <-ch
				if !ok { // if there is nothing to do and the channel has been closed then end the goroutine
					wg.Done()
					return
				}
				mu.Lock()
				results = append(results, job(a[0], a[1], a[2]))
				mu.Unlock()
			}
		}()
	}

	// Now the jobs can be added to the channel, which is used as a queue
	for _, artifact := range artifacts {
		ch <- artifact // add artifact to the queue
	}

	close(ch) // This tells the goroutines there's nothing else to do
	wg.Wait() // Wait for the threads to finish

	finalResult := true
	for _, result := range results {
		if finalResult = result; !finalResult {
			break
		}
	}

	return finalResult
}
