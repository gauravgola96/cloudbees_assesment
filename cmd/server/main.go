package main

import (
	"context"
	"github.com/gauravgola96/cloudbees_assesment/pkg/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	srv, err := server.NewHttpServer()
	if err != nil {
		log.Fatalf("error starting server: %s", err.Error())
		return
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	sig := <-quit
	log.Printf("server shutdown due to %s", sig.String())
	err = srv.Shutdown(context.TODO())
	if err != nil {
		log.Fatalf("error server gracefull %s", err.Error())
		return
	}
	log.Printf("server stopped gracefully")
}
