// Package postgres opens the database/sql connection used by the storage
// services. Persistence itself is done through the shop-db-models library
// (github.com/khetzl/shop-db-models/pkg/storage/dbmodel), whose SQLBoiler
// models are generated for the lib/pq driver.
//
// The database schema is owned by the shop-db-models repository
// (db-scripts/0000_init_up.sql, applied via its `make push_model`); this
// service only connects to an already-provisioned database and does not create
// or migrate any tables of its own.
package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // register the "postgres" database/sql driver
)

// Connect opens a database/sql pool for the given DSN and verifies
// connectivity with a ping.
func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: open: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return db, nil
}
