package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/zeny-io/mboxparser"
)

// Flags encapsulate command-line parameters
type Flags struct {
	file      string
	threshold int
}

var flags Flags

func initFlags() {

	flag.StringVar(&flags.file, "file", "sample.mbox", "mailbox to parse")
	flag.IntVar(&flags.threshold, "threshold", 3, "number of mails to use as threshold")

	flag.Parse()
}

func main() {
	initFlags()

	mailMap := make(map[string]int)
	mu := &sync.Mutex{}

	mbox, err := mboxparser.ReadFile(flags.file)
	if err != nil {
		os.Exit(2)
	}

	var wg sync.WaitGroup
	numCpus := runtime.NumCPU()
	mailBoxLen := len(mbox.Messages)

	wg.Add(numCpus)
	for i := 0; i < numCpus; i++ {
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()

			startIndex := i * (mailBoxLen / numCpus)
			if i >= mailBoxLen {
				return
			}

			endIndex := (i + 1) * ((mailBoxLen / numCpus) + 1)
			if endIndex >= mailBoxLen {
				endIndex = mailBoxLen
			}

			for _, mail := range mbox.Messages[startIndex:endIndex] {
				for k, vs := range mail.Header {
					if k == "From" {
						mu.Lock()
						if _, ok := mailMap[vs[0]]; !ok {
							mailMap[vs[0]] = 0
						}
						mailMap[vs[0]]++
						mu.Unlock()
					}
				}
			}
		}(i, &wg)
	}
	wg.Wait()

	for k, v := range mailMap {
		if v >= flags.threshold {
			fmt.Printf("%s: %d\n", k, v)
		}
	}
}
