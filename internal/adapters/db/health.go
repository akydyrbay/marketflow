package db

func (repo *PostgresRepository) CheckHealth() error {
	if err := repo.db.Ping(); err != nil {
		return err
	}
	return nil
}
