package app

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	m "marketflow/internal/domain"
)

func FanOut(ctx context.Context, in <-chan m.PriceUpdate, count int) []<-chan m.PriceUpdate {
	outs := make([]chan m.PriceUpdate, count)
	for i := 0; i < count; i++ {
		outs[i] = make(chan m.PriceUpdate)
		go func(out chan m.PriceUpdate) {
			defer close(out)
			for {
				select {
				case <-ctx.Done():
					return
				case update, ok := <-in:
					if !ok {
						return
					}
					out <- update
				}
			}
		}(outs[i])
	}

	routs := make([]<-chan m.PriceUpdate, count)
	for i, ch := range outs {
		routs[i] = ch
	}
	return routs
}

func ReadData(file string) <-chan string {
	f, err := os.Open(file) // opens the file for reading
	if err != nil {
		log.Fatal(err)
	}

	out := make(chan string) // channel declared

	// returns a scanner to read from f
	fileScanner := bufio.NewScanner(f)
	fileScanner.Split(bufio.ScanLines) // scanning it line-by-line token

	// loop through the fileScanner based on our token split
	go func() {
		for fileScanner.Scan() {
			val := fileScanner.Text() // returns the recent token
			out <- val                // passed the token value to our channel
		}

		close(out) // closed the channel when all content of file is read

		// closed the file
		err := f.Close()
		if err != nil {
			fmt.Printf("Unable to close an opened file: %v\n", err.Error())
			return
		}
	}()

	return out
}

func FanIn(ctx context.Context, chans []<-chan m.PriceUpdate) <-chan m.PriceUpdate {
	out := make(chan m.PriceUpdate)
	var wg sync.WaitGroup
	wg.Add(len(chans))

	for _, ch := range chans {
		go func(c <-chan m.PriceUpdate) {
			defer wg.Done()
			for update := range c {
				select {
				case <-ctx.Done():
					return
				case out <- update:

				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
