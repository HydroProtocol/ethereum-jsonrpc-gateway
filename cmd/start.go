package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HydroProtocol/ethereum-jsonrpc-gateway/core"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use: "start",
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(Run())
	},
}

func waitExitSignal(ctxStop context.CancelFunc) {
	var exitSignal = make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGTERM)
	signal.Notify(exitSignal, syscall.SIGINT)

	<-exitSignal

	logrus.Info("Stopping...")
	ctxStop()
}

func Run() int {

	ctx, stop := context.WithCancel(context.Background())
	go waitExitSignal(stop)

	err := core.LoadConfig(ctx)

	if err != nil {
		logrus.Fatal(err)
	}

	httpServer := &http.Server{Addr: ":3005", Handler: &core.Server{}}

	// http server graceful shutdown
	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logrus.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
	}()

	logrus.Infof("Listening on http://0.0.0.0%s\n", httpServer.Addr)

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		logrus.Fatal(err)
	}

	logrus.Info("Stopped")
	return 0
}
