package handler

import (
	"github.com/askthechain/atcutils/pkg/logger"

	"github.com/khetzl/shop-customer-api/internal/domain/models"
)

//go:generate mockery --output=./tests/mocks --name=ItemStore
type ItemStore interface {
	GetAll() ([]models.Item, error)
}

type Handler struct {
	logger    atclogger.Logger
	itemStore ItemStore
}

func New(
	logger atclogger.Logger,
	itemStore ItemStore,
) (*Handler, error) {
	return &Handler{
		logger:    logger,
		itemStore: itemStore,
	}, nil
}
