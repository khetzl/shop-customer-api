package bootstrap

import (
	"database/sql"
	"net/http"

	"github.com/askthechain/atcutils/pkg/logger"

	"github.com/khetzl/shop-customer-api/cmd/api/config"
	"github.com/khetzl/shop-customer-api/internal/api/handler"
	"github.com/khetzl/shop-customer-api/internal/api/router"
	"github.com/khetzl/shop-customer-api/internal/api/server"
	"github.com/khetzl/shop-customer-api/internal/services/itemstore"
	"github.com/khetzl/shop-customer-api/internal/storage/postgres"
)

func NewLogger() atclogger.Logger {
	logger, err := atclogger.NewSimpleConsoleLogger(atclogger.Info)
	if err != nil {
		panic(err)
	}
	return logger
}

func NewAPI(logger atclogger.Logger) *server.Service {
	db := newDB(logger)

	itemStore := newItemStore(logger, db)
	server := newWebservice(logger, itemStore)
	return server
}

// newDB opens the shared PostgreSQL connection. The schema is provisioned from
// the shop-db-models repository (db-scripts/0000_init_up.sql); this service
// only connects to it.
func newDB(logger atclogger.Logger) *sql.DB {
	db, err := postgres.Connect(config.Cfg.Database.URL)
	if err != nil {
		panic(err)
	}
	return db
}

func newWebservice(
	logger atclogger.Logger,
	itemStore *itemstore.Service,
) *server.Service {
	handler, err := handler.New(logger, itemStore)
	if err != nil {
		panic(err)
	}

	router, err := router.New(logger, handler)
	if err != nil {
		panic(err)
	}

	httpServer := &http.Server{
		Addr:    config.Cfg.API.Address,
		Handler: router,
	}
	server, err := server.New(logger, httpServer)
	if err != nil {
		panic(err)
	}
	return server
}

func newItemStore(logger atclogger.Logger, db *sql.DB) *itemstore.Service {
	itemStore, err := itemstore.New(logger, db)
	if err != nil {
		panic(err)
	}
	return itemStore
}
