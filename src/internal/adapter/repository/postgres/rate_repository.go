package postgres

import "database/sql"

type RateRepository struct {
	db *sql.DB
}

func NewRateRepository(db *sql.DB) *RateRepository {
	return &RateRepository{db: db}
}
