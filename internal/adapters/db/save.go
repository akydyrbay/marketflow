package db

import "marketflow/internal/domain"

func (repo *PostgresRepository) SaveLatestData(latestData map[string]domain.Data) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO LatestData (Exchange, Pair_name, Price, StoredTime)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (Exchange, Pair_name) DO UPDATE
		SET Price = EXCLUDED.Price,
    	StoredTime = EXCLUDED.StoredTime;
		`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, data := range latestData {
		if _, err := stmt.Exec(data.ExchangeName, data.Symbol, data.Price, data.Timestamp); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
