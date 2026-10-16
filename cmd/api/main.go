package main

import (
	"github.com/khetzl/shop-customer-api/cmd/api/bootstrap"
	"github.com/khetzl/shop-customer-api/cmd/api/config"

	"github.com/askthechain/atcutils/pkg/process"
)

func main() {

	err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := bootstrap.NewLogger()
	api := bootstrap.NewAPI(logger)

	process.Run(logger, api)
}
