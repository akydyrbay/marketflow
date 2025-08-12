package db

import (
	"time"

	"marketflow/internal/domain"
	"marketflow/pkg/logger"
)

// Gets the latest price data by exchange for specific symbol
func (repo *PostgresRepository) LatestDataByExchange(exchange, symbol string) (domain.Data, error) {
	data := domain.Data{
		ExchangeName: exchange,
		Symbol:       symbol,
	}

	rows, err := repo.db.Query(`
		SELECT Exchange, Pair_name, Price, StoredTime
			FROM LatestData
		WHERE Exchange = $1 AND Pair_name = $2
		ORDER BY StoredTime DESC
		LIMIT 1;
		`, exchange, symbol)
	if err != nil {
		return domain.Data{}, err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&data.ExchangeName, &data.Symbol, &data.Price, &data.Timestamp); err != nil {
			return domain.Data{}, err
		}

		return data, nil
	}

	return domain.Data{}, nil
}

func (repo *PostgresRepository) LatestDataByAllExchanges(symbol string) (domain.Data, error) {
	data := domain.Data{
		ExchangeName: "All",
		Symbol:       symbol,
	}

	rows, err := repo.db.Query(`
		SELECT Exchange, Pair_name, Price, StoredTime
		FROM LatestData
		WHERE Pair_name = $1
		ORDER BY StoredTime DESC
		LIMIT 1;
	`, symbol)
	if err != nil {
		return domain.Data{}, err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&data.ExchangeName, &data.Symbol, &data.Price, &data.Timestamp); err != nil {
			return domain.Data{}, err
		}
		return data, nil
	}

	return domain.Data{}, nil
}

// Gets the average price data by exchange over all period
func (repo *PostgresRepository) AveragePriceByExchange(exchange, symbol string) (domain.Data, error) {
	data := domain.Data{
		ExchangeName: exchange,
		Symbol:       symbol,
	}

	rows, err := repo.db.Query(`
	SELECT COALESCE(AVG(Average_price), 0) FROM AggregatedData
	WHERE Exchange = $1 AND Pair_name = $2
	`, exchange, symbol)
	if err != nil {
		return domain.Data{}, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&data.Price); err != nil {
			return domain.Data{}, err
		}
	}

	return data, nil
}

// Gets the average price by exchange over all period
func (repo *PostgresRepository) AveragePriceByAllExchanges(symbol string) (domain.Data, error) {
	data := domain.Data{
		ExchangeName: "All",
		Symbol:       symbol,
	}

	rows, err := repo.db.Query(`
	SELECT COALESCE(AVG(Average_price), 0) from AggregatedData
	WHERE Pair_name = $1 AND Exchange = 'All'
	`, symbol)
	if err != nil {
		return domain.Data{}, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&data.Price); err != nil {
			return domain.Data{}, err
		}
	}
	logger.Debug("Debuger: ", data)
	return data, nil
}

// Gets the average price within the last {duration}
func (repo *PostgresRepository) AveragePriceWithDuration(exchange, symbol string, startTime time.Time, duration time.Duration) (domain.Data, error) {
	data := domain.Data{
		ExchangeName: exchange,
		Symbol:       symbol,
	}

	rows, err := repo.db.Query(`
	SELECT COALESCE(AVG(Average_price), 0) FROM AggregatedData
	WHERE Exchange = $1 AND Pair_name = $2 AND StoredTime BETWEEN $3 and $4
	`, exchange, symbol, startTime.Add(-duration), startTime)
	if err != nil {
		return domain.Data{}, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&data.Price); err != nil {
			return domain.Data{}, err
		}
	}

	return data, nil
}

// Min by all exchange and all time
func (repo *PostgresRepository) MinPriceByAllExchanges(symbol string) (domain.Data, error) {
	data := domain.Data{
		ExchangeName: "All",
		Symbol:       symbol,
	}

	rows, err := repo.db.Query(`
SELECT Pair_name, exchange, StoredTime, Min_price
FROM AggregatedData
WHERE 
    Pair_name = $1  AND exchange = 'All'
    AND Min_price = (
        SELECT MIN(Min_price)
        FROM AggregatedData
        WHERE 
            Pair_name = $1
            AND exchange = 'All'
    );
	`, symbol)
	if err != nil {
		return domain.Data{}, err
	}
	defer rows.Close()

	var t time.Time
	for rows.Next() {
		if err := rows.Scan(&data.Symbol, &data.ExchangeName, &t, &data.Price); err != nil {
			return domain.Data{}, err
		}
	}
	data.Timestamp = t.UnixMilli()

	return data, nil
}

// Min by one exchange and all time
func (repo *PostgresRepository) MinPriceByExchange(exchange, symbol string) (domain.Data, error) {
	data := domain.Data{
		ExchangeName: exchange,
		Symbol:       symbol,
	}

	rows, err := repo.db.Query(`
SELECT Pair_name, exchange, StoredTime, Min_price
FROM AggregatedData
WHERE 
    Pair_name = $1  AND exchange = $2
    AND Min_price = (
        SELECT MIN(Min_price)
        FROM AggregatedData
        WHERE 
            Pair_name = $1
            AND exchange = $2
    );
	`, exchange, symbol)
	if err != nil {
		return domain.Data{}, err
	}
	defer rows.Close()

	var t time.Time
	for rows.Next() {
		if err := rows.Scan(&data.Symbol, &data.ExchangeName, &t, &data.Price); err != nil {
			return domain.Data{}, err
		}
	}
	data.Timestamp = t.UnixMilli()

	return data, nil
}

// Min by one exchange on period
func (repo *PostgresRepository) MinPriceByExchangeWithDuration(exchange, symbol string, startTime time.Time, duration time.Duration) (domain.Data, error) {
	data := domain.Data{
		ExchangeName: exchange,
		Symbol:       symbol,
	}

	rows, err := repo.db.Query(`
SELECT Pair_name, exchange, StoredTime, Min_price
FROM AggregatedData
WHERE 
    Pair_name = $1  AND exchange =  $2 AND StoredTime BETWEEN $3 AND $4
    AND Min_price = (
        SELECT MIN(Min_price)
        FROM AggregatedData
        WHERE 
            Pair_name = $1
            AND exchange = $2
            AND StoredTime BETWEEN $3 AND $4
    );
	`, symbol, exchange, startTime.Add(-duration), startTime)
	if err != nil {
		return domain.Data{}, err
	}
	defer rows.Close()

	var t time.Time
	for rows.Next() {
		if err := rows.Scan(&data.Symbol, &data.ExchangeName, &t, &data.Price); err != nil {
			return domain.Data{}, err
		}
	}

	data.Timestamp = t.UnixMilli()

	return data, nil
}

// Min by one exchange on period
func (repo *PostgresRepository) MinPriceByAllExchangesWithDuration(symbol string, startTime time.Time, duration time.Duration) (domain.Data, error) {
	data := domain.Data{
		ExchangeName: "All",
		Symbol:       symbol,
	}

	rows, err := repo.db.Query(`
SELECT Pair_name, exchange, StoredTime, Min_price
FROM AggregatedData
WHERE 
    Pair_name = $1  AND exchange = 'All' AND StoredTime BETWEEN $2 AND $3
    AND Min_price = (
        SELECT MIN(Min_price)
        FROM AggregatedData
        WHERE 
            Pair_name = $1
            AND exchange = 'All'
            AND StoredTime BETWEEN $2 AND $3
    );
	`, symbol, startTime.Add(-duration), startTime)
	if err != nil {
		return domain.Data{}, err
	}
	defer rows.Close()

	var t time.Time
	for rows.Next() {
		if err := rows.Scan(&data.Symbol, &data.ExchangeName, &t, &data.Price); err != nil {
			return domain.Data{}, err
		}
	}
	data.Timestamp = t.UnixMilli()

	return data, nil
}

// Max by all exchange all time
func (repo *PostgresRepository) MaxPriceByAllExchanges(symbol string) (domain.Data, error) {
	data := domain.Data{
		ExchangeName: "All",
		Symbol:       symbol,
	}

	rows, err := repo.db.Query(`
SELECT Pair_name, exchange, StoredTime, Max_price
FROM AggregatedData
WHERE 
    Pair_name = $1  AND exchange = 'All'
    AND Max_price = (
        SELECT MAX(Max_price)
        FROM AggregatedData
        WHERE 
            Pair_name = $1
            AND exchange = 'All'
    );

	`, symbol)
	if err != nil {
		return domain.Data{}, err
	}
	defer rows.Close()

	var t time.Time
	for rows.Next() {
		if err := rows.Scan(&data.Symbol, &data.ExchangeName, &t, &data.Price); err != nil {
			return domain.Data{}, err
		}
	}
	data.Timestamp = t.UnixMilli()

	return data, nil
}

// Max by one exchange on all time
func (repo *PostgresRepository) MaxPriceByExchange(exchange, symbol string) (domain.Data, error) {
	data := domain.Data{
		ExchangeName: exchange,
		Symbol:       symbol,
	}

	rows, err := repo.db.Query(`
SELECT Pair_name, exchange, StoredTime, Max_price
FROM AggregatedData
WHERE 
    Pair_name = $1  AND exchange = $2
    AND Max_price = (
        SELECT MAX(Max_price)
        FROM AggregatedData
        WHERE 
            Pair_name = $1
            AND exchange = $2
    );
	`, symbol, exchange)
	if err != nil {
		return domain.Data{}, err
	}
	defer rows.Close()

	var t time.Time
	for rows.Next() {
		if err := rows.Scan(&data.Symbol, &data.ExchangeName, &t, &data.Price); err != nil {
			return domain.Data{}, err
		}
	}
	data.Timestamp = t.UnixMilli()

	return data, nil
}

// Max by one exchange on period
func (repo *PostgresRepository) MaxPriceByExchangeWithDuration(exchange, symbol string, startTime time.Time, duration time.Duration) (domain.Data, error) {
	data := domain.Data{
		ExchangeName: exchange,
		Symbol:       symbol,
	}

	rows, err := repo.db.Query(`
SELECT Pair_name, exchange, StoredTime, Max_price
FROM AggregatedData
WHERE 
    Pair_name = $1  AND exchange = $2 AND StoredTime BETWEEN $3 AND $4
    AND Max_price = (
        SELECT MAX(Max_price)
        FROM AggregatedData
        WHERE 
            Pair_name = $1
            AND exchange = $2
            AND StoredTime BETWEEN $3 AND $4
    );
	`, symbol, exchange, startTime.Add(-duration), startTime)
	if err != nil {
		return domain.Data{}, err
	}
	defer rows.Close()

	var t time.Time
	for rows.Next() {
		if err := rows.Scan(&data.Symbol, &data.ExchangeName, &t, &data.Price); err != nil {
			return domain.Data{}, err
		}
	}

	data.Timestamp = t.UnixMilli()

	return data, nil
}

// Max by all exchange on period
func (repo *PostgresRepository) MaxPriceByAllExchangesWithDuration(symbol string, startTime time.Time, duration time.Duration) (domain.Data, error) {
	data := domain.Data{
		ExchangeName: "All",
		Symbol:       symbol,
	}

	rows, err := repo.db.Query(`
SELECT Pair_name, exchange, StoredTime, Max_price
FROM AggregatedData
WHERE 
    Pair_name = $1  AND exchange = 'All' AND StoredTime BETWEEN $2 AND $3
    AND Max_price = (
        SELECT MAX(Max_price)
        FROM AggregatedData
        WHERE 
            Pair_name = $1
            AND exchange = 'All'
            AND StoredTime BETWEEN $2 AND $3
    );

	`, symbol, startTime.Add(-duration), startTime)
	if err != nil {
		return domain.Data{}, err
	}
	defer rows.Close()

	var t time.Time
	for rows.Next() {
		if err := rows.Scan(&data.Symbol, &data.ExchangeName, &t, &data.Price); err != nil {
			return domain.Data{}, err
		}
	}

	data.Timestamp = t.UnixMilli()

	return data, nil
}
