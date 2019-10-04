package main

import (
	"context"
	"github.com/luca-moser/iota-bounty-platform/server/controllers"
	"github.com/luca-moser/iota-bounty-platform/server/server"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	srv := server.Server{}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	srv.Start()
	select {
	case <-sigs:
		srv.Logger.Info("interrupt signal...")
		ctx, _ := context.WithTimeout(context.Background(), time.Duration(1500)*time.Millisecond)
		srv.Shutdown(ctx)
		controllers.ShutdownWebHookListener()
	}

}
