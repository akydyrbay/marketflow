package pipeline

import (
	"marketflow/internal/domain"
	"marketflow/pkg/logger"
	"sync"
	"time"
)

func FanIn(dataFlows [3]chan domain.Data) chan []domain.Data {
	mergedCh := make(chan domain.Data, 15)
	ch := make(chan []domain.Data, 3)

	closedCount := 0
	var muClosed sync.Mutex

	go func() {
		defer close(mergedCh)

	mainLoop:
		for {
			select {
			case e1, ok := <-dataFlows[0]:
				if !ok {
					muClosed.Lock()
					closedCount++
					muClosed.Unlock()
					dataFlows[0] = nil
				} else {
					mergedCh <- e1
				}
			case e2, ok := <-dataFlows[1]:
				if !ok {
					muClosed.Lock()
					closedCount++
					muClosed.Unlock()
					dataFlows[1] = nil
				} else {
					mergedCh <- e2
				}
			case e3, ok := <-dataFlows[2]:
				if !ok {
					muClosed.Lock()
					closedCount++
					muClosed.Unlock()
					dataFlows[2] = nil
				} else {
					mergedCh <- e3
				}
			}

			muClosed.Lock()
			if closedCount == 3 {
				muClosed.Unlock()
				break mainLoop
			}
			muClosed.Unlock()
		}
	}()

	t := time.NewTicker(time.Second)
	rawData := make([]domain.Data, 0)
	done := make(chan bool)
	mu := sync.Mutex{}

	go func() {
		defer close(ch)

	mainLoop:
		for {
			select {
			case tick := <-t.C:
				logger.Debug(tick.String())
				mu.Lock()

				if len(rawData) == 0 {
					mu.Unlock()
					continue
				}
				ch <- rawData
				rawData = make([]domain.Data, 0)

				mu.Unlock()
			case <-done:
				break mainLoop
			}
		}
	}()

	go func() {
		for data := range mergedCh {
			mu.Lock()
			rawData = append(rawData, data)
			mu.Unlock()
		}

		done <- true
		close(done)
		t.Stop()
	}()

	return ch
}
