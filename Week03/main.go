package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	g := errgroup.Group{}
	ctx, cancelFunc := context.WithCancel(context.Background())

	g.Go(func() error {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		select {
		case sig := <-sigChan:
			cancelFunc()
			return fmt.Errorf("singal: %v", sig)
		case <-ctx.Done():
			return nil
		}
	})

	g.Go(func() error {
		serverMux := http.NewServeMux()
		serverMux.HandleFunc("/", rootHandler)
		server := http.Server{
			Addr:    ":8888",
			Handler: serverMux,
		}

		go func() {
			defer server.Shutdown(context.TODO())
			<-ctx.Done()
			fmt.Println("server stop")
		}()

		return server.ListenAndServe()
	})

	if err := g.Wait(); err != nil {
		log.Fatalf("error: %s", err)
	}
}

func rootHandler(res http.ResponseWriter, req *http.Request) {
	time.Sleep(time.Microsecond * time.Duration(rand.Intn(10)))
}
