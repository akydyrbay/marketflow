package pipeline

import (
	"encoding/json"
	"marketflow/internal/domain"
	"marketflow/pkg/logger"
	"sync"
)

// Worker processes tasks from the jobs channel and sends the results to the results channel
func Worker(number string, jobs chan string, results chan domain.Data, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		data := domain.Data{}
		err := json.Unmarshal([]byte(j), &data)
		if err != nil {
			logger.Error("Unmarshalling error in worker", err.Error())
			continue
		}

		// Assign the name of the exchange and send it to the results channel
		data.ExchangeName = number
		results <- data
	}
}
