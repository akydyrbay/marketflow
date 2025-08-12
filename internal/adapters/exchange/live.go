package exchange

import (
	"bufio"
	"errors"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"marketflow/internal/domain"
	"marketflow/internal/service"
	"marketflow/pkg/logger"
)

type Exchange struct {
	number      string
	conn        net.Conn
	closeCh     chan bool
	messageChan chan string
}

type LiveMode struct {
	Exchanges []*Exchange
	mu        sync.Mutex
}

func NewLiveModeFetcher() *LiveMode {
	return &LiveMode{Exchanges: make([]*Exchange, 0)}
}

func (m *LiveMode) SetupDataFetcher() (chan map[string]domain.ExchangeData, chan []domain.Data, error) {
	dataFlows := [3]chan domain.Data{make(chan domain.Data), make(chan domain.Data), make(chan domain.Data)}

	wg := &sync.WaitGroup{}

	ports := []string{os.Getenv("EXCHANGE1_PORT"), os.Getenv("EXCHANGE2_PORT"), os.Getenv("EXCHANGE3_PORT")}
	exchHosts := []string{os.Getenv("EXCHANGE1_NAME"), os.Getenv("EXCHANGE2_NAME"), os.Getenv("EXCHANGE3_NAME")}

	for i := 0; i < len(ports); i++ {
		wg.Add(1)
		exch, err := GenerateExchange("Exchange"+strconv.Itoa(i+1), exchHosts[i]+":"+ports[i])
		if err != nil {
			logger.Error("Failed to connect exchange number: %d, error: %s", i+1, err.Error())
			wg.Done()
			continue
		}

		// Receive data from the server
		go exch.FetchData(wg)

		// Start the worker to process the received data
		go exch.SetWorkers(wg, dataFlows[i])

		m.Exchanges = append(m.Exchanges, exch)
	}

	if len(m.Exchanges) != 3 {
		return nil, nil, errors.New("failed to connect to 3 exchanges")
	}

	mergedCh := service.FanIn(dataFlows)

	aggregated, rawDatach := service.Aggregate(mergedCh)

	go func() {
		wg.Wait()
		for i := 0; i < len(m.Exchanges); i++ {
			if m.Exchanges[i] == nil {
				continue
			}
			if m.Exchanges[i].conn != nil {
				m.Exchanges[i].conn.Close()
			}
		}

		logger.Info("All workers have finished processing.")
	}()
	return aggregated, rawDatach, nil
}

// GenerateExchange returns pointer to Exchange data with messageChan
func GenerateExchange(number string, address string) (*Exchange, error) {
	messageChan := make(chan string)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	exchangeServ := &Exchange{number: number, conn: conn, messageChan: messageChan}
	return exchangeServ, nil
}

// SetWorkers starts goroutine workers to process data
func (exch *Exchange) SetWorkers(globalWg *sync.WaitGroup, fan_in chan domain.Data) {
	workerWg := &sync.WaitGroup{}
	for w := 1; w <= 5; w++ {
		workerWg.Add(1)
		globalWg.Add(1)
		go func() {
			service.Worker(exch.number, exch.messageChan, fan_in, workerWg)
			globalWg.Done()
		}()
	}

	go func() {
		workerWg.Wait()
		logger.Debug("Local workers finished work in exchange ", exch.number)
		close(fan_in)
	}()
}

func (exch *Exchange) FetchData(wg *sync.WaitGroup) {
	defer wg.Done()

	scanner := bufio.NewScanner(exch.conn)
	address := exch.conn.RemoteAddr().String()

	closeCh := make(chan bool, 1)
	exch.closeCh = closeCh

	reconnect := true
	mu := sync.Mutex{}

	go func() {
		<-closeCh
		mu.Lock()
		reconnect = false
		mu.Unlock()
	}()

	logger.Info("Starting reading data on exchange: ", exch.number)

	for {
		for scanner.Scan() && reconnect {
			line := scanner.Text()
			exch.messageChan <- line
		}

		logger.Info("Connection lost on exchange %s. Reconnecting...\n", exch.number)

		if reconnect {
			if err := exch.Reconnect(address); err != nil {
				logger.Error("Failed to reconnect exchange %s: %v", exch.number, err)
				break
			}

			scanner = bufio.NewScanner(exch.conn)
		} else {
			break
		}
	}

	logger.Info("Giving up on exchange: ", exch.number)
	close(closeCh)
	close(exch.messageChan)
}

func (exch *Exchange) Reconnect(address string) error {
	var err error
	for i := 0; i < 5; i++ {
		time.Sleep(2 * time.Second)
		exch.conn, err = net.Dial("tcp", address)
		if err == nil {
			logger.Info("Reconnected to exchange: ", exch.number)
			return nil
		}
		logger.Warn("Reconnect attempt failed: ", err)
	}
	return err
}

func (m *LiveMode) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := 0; i < len(m.Exchanges); i++ {
		if m.Exchanges[i] == nil || m.Exchanges[i].conn == nil {
			continue
		}

		if err := m.Exchanges[i].conn.Close(); err != nil {
			logger.Error("Failed to close connection: ", err)
			continue
		}

		m.Exchanges[i].closeCh <- true
	}
}

func (m *LiveMode) CheckHealth() error {
	var unhealthy string
	for i := 0; i < len(m.Exchanges); i++ {
		select {
		case _, ok := <-m.Exchanges[i].messageChan:
			if !ok {
				unhealthy += m.Exchanges[i].number + " "
			}
		default:
			continue
		}
	}
	if len(unhealthy) != 0 {
		return errors.New("unhealthy exchanges: " + unhealthy)
	}
	return nil
}
