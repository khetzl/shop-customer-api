package server

import (
	"context"
	"time"

	"github.com/askthechain/atcutils/pkg/logger"
	"github.com/askthechain/atcutils/pkg/process"
)

type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

type Service struct {
	logger     atclogger.Logger
	httpServer HTTPServer
}

func New(logger atclogger.Logger,
	httpServer HTTPServer,
) (*Service, error) {
	return &Service{
		logger:     logger,
		httpServer: httpServer,
	}, nil
}

func (g *Service) Start(ctx context.Context) error {
	err := make(chan error)

	// Start serving in a separate goroutine
	go func() {
		defer process.CapturePanic(g.logger, false, nil)
		e := g.httpServer.ListenAndServe()
		err <- e
	}()

	select {
	case e := <-err:
		return e
	case <-ctx.Done():
		shutdownTimeout := time.Second
		doneCtx, ignorable := context.WithTimeout(context.Background(), shutdownTimeout)
		defer ignorable()

		err := g.httpServer.Shutdown(doneCtx)
		if err != nil {
			g.logger.Error("webservice server termination error", atclogger.Params{"err": err})
			return err
		}
	}

	return nil
}
