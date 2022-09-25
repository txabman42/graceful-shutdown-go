package graceful_shutdown_go

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

type GracefulAction interface {
	Start() error
	Stop(ctx context.Context) error
}

type Level uint16

const (
	defaultStartTimeout    = 10 * time.Second
	defaultShutdownTimeout = 10 * time.Second

	LOW Level = iota
	MID
	HIGH
)

var sortedLevels = []Level{HIGH, MID, LOW}

type GracefulShutdown struct {
	numActions uint16
	phases     map[Level][]GracefulAction
}

func NewGracefulShutdown() *GracefulShutdown {
	return &GracefulShutdown{
		phases: map[Level][]GracefulAction{},
	}
}

func (gs *GracefulShutdown) Register(a GracefulAction, l Level) {
	done := make(chan interface{})
	ctx, cancel := context.WithTimeout(context.Background(), defaultStartTimeout)
	defer cancel()

	go func() {
		if err := a.Start(); err != nil {
			log.Panic(err)
		}
		done <- true
	}()

	select {
	case <-ctx.Done():
		log.Panicf("action %T had not enough time to start correctly in %s seconds", a, defaultStartTimeout)
	case <-done:
	}

	gs.numActions++
	gs.phases[l] = append(gs.phases[l], a)
}

func (gs *GracefulShutdown) Run() {
	s := make(chan os.Signal, 1)
	// Wait until any of this os.Signals is thrown
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-s

	gs.gracefulShutdown()
}

func (gs *GracefulShutdown) gracefulShutdown() {
	log.Info("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	for _, level := range sortedLevels {
		for _, a := range gs.phases[level] {
			select {
			case <-ctx.Done():
				log.Error("Shutting down process was cancelled as shutdown timeout was exceeded")
				return
			default:
				if err := a.Stop(ctx); err != nil {
					log.Error("Shutting down process failed: %s", err)
					return
				}
			}
		}
	}
}
