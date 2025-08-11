package db

import (
	"marketflow/internal/domain"
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
