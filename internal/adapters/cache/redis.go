package cache

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
)

type RedisCache struct {
	Addr   string
	conn   net.Conn
	reader *bufio.Reader
}

func (r *RedisCache) Connect() error {
	conn, err := net.Dial("tcp", r.Addr)
	if err != nil {
		return fmt.Errorf("failed to connect to Redis at %s: %w", r.Addr, err)
	}
	r.conn = conn
	r.reader = bufio.NewReader(conn)
	return nil
}

// sendCmd sends a raw inline Redis command and returns the first line of the reply.
func (r *RedisCache) sendCmd(cmd string) (string, error) {
	// Commands must end with \r\n
	if _, err := r.conn.Write([]byte(cmd + "\r\n")); err != nil {
		return "", fmt.Errorf("write error: %w", err)
	}
	// Read the first line of the response
	line, err := r.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read error: %w", err)
	}
	return line, nil
}

// AddPrice adds a price update to a sorted set keyed by exchange and symbol.
// It uses the timestamp as the score and the price as the member.
func (r *RedisCache) AddPrice(exchange, symbol string, price float64, timestamp int64) error {
	key := fmt.Sprintf("prices:%s:%s", exchange, symbol)
	score := strconv.FormatInt(timestamp, 10)
	member := strconv.FormatFloat(price, 'f', -1, 64)
	cmd := fmt.Sprintf("ZADD %s %s %s", key, score, member)
	_, err := r.sendCmd(cmd)
	return err
}

// Cleanup removes entries older than the given cutoff timestamp (exclusive).
func (r *RedisCache) Cleanup(exchange, symbol string, cutoff int64) error {
	key := fmt.Sprintf("prices:%s:%s", exchange, symbol)
	cutoffStr := strconv.FormatInt(cutoff, 10)
	// Remove scores from -inf to cutoff
	cmd := fmt.Sprintf("ZREMRANGEBYSCORE %s -inf %s", key, cutoffStr)
	_, err := r.sendCmd(cmd)
	return err
}
